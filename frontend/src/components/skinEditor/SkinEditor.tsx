import { useEffect, useMemo, useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { useSkinEditor } from "../../hooks/useSkinEditor";
import { Command } from "../../models/command";
import SkinModel, { CSSRule, PreviewUpdate } from "../../models/skinModel";
import { CompareAgainst, Comparison } from "../splitter/Splitter";
import { mergeSkinElements } from "./mergeSkinElements";
import { previewConfig } from "./previewBase";
import { previewSession as completedPreview } from "./previewCompletedSession";
import { previewElements } from "./previewElements";
import { previewSession as emptyPreview } from "./previewEmptySession";
import { previewSession as inProgressPreview } from "./previewInProgressSession";
import PreviewSplitter from "./PreviewSplitter";
import { registerPreviewElements } from "./registerPreviewElements";
import { collectRuntimeCSS } from "./runtimeCSSCollector";
import SkinControls from "./SkinControls";
import SkinFiles from "./SkinFiles";

interface Props {
    model: SkinModel;
}

export default function SkinEditor({ model }: Props) {
    const [preview, setPreview] = useState<"empty" | "running" | "completed">("running");
    const [comparison, setComparison] = useState<Comparison>(CompareAgainst.Average);

    const session = useMemo(() => {
        switch (preview) {
            case "empty":
                return emptyPreview;
            case "completed":
                return completedPreview;
            default:
                return inProgressPreview;
        }
    }, [preview]);

    useEffect(() => {
        registerPreviewElements(previewElements);
    }, []);

    const editor = useSkinEditor(model);

    const runtimeRules = useMemo<CSSRule[]>(() => collectRuntimeCSS(), []);

    const [previewUpdate, setPreviewUpdate] = useState<PreviewUpdate>({
        elements: [],
        metrics: {
            splitter: new DOMRect(),

            content: new DOMRect(),

            canvasWidth: 0,

            canvasHeight: 0,

            paddingLeft: 0,

            paddingRight: 0,

            paddingTop: 0,

            paddingBottom: 0,

            overflowX: 0,

            overflowY: 0,

            splitterOffsetX: 0,

            splitterOffsetY: 0,

            hasCanvasOverflow: false,

            hasElementOverflow: false,

            overflowingElements: [],
        },
    });

    const runtimeElements = previewUpdate.elements;

    const hasPreviewOverflow = previewUpdate.metrics.hasElementOverflow;

    const allElements = useMemo(
        () => mergeSkinElements(previewElements, model.elements, runtimeElements),
        [model.elements, runtimeElements],
    );

    const availableElements = useMemo(
        () => mergeSkinElements(previewElements, model.elements, runtimeElements),
        [model.elements, runtimeElements],
    );

    const selectedRuntime = editor.target.elementId
        ? (runtimeElements.find((element) => element.id === editor.target.elementId) ?? null)
        : null;

    const availableIds = useMemo(() => new Set(runtimeElements.map((element) => element.id)), [runtimeElements]);

    const combinedModel = useMemo<SkinModel>(
        () => ({
            ...model,
            rules: [...editor.flatRules, ...runtimeRules],
        }),
        [model, runtimeRules],
    );

    /**
     * The preview should render the backend working copy, not the
     * installed files on disk.
     */
    const overrideCSS = useMemo(() => {
        return editor.flatRules
            .filter((rule) => rule.selector)
            .map(
                (rule) => `
     ${rule.selector} {
     ${rule.body}
     }
     `,
            )
            .join("\n");
    }, [editor.flatRules]);

    async function cancel() {
        await Dispatch(Command.CANCEL, null);
    }

    return (
        <div className="skin-editor">
            <section className="skin-editor-left panel skin-editor-column">
                <SkinControls
                    model={combinedModel}
                    elements={allElements}
                    availableIds={availableIds}
                    selectedElement={editor.target.elementId}
                    selectedFile={editor.target.file}
                    activeRule={editor.activeRule}
                    onElementSelected={editor.selectElement}
                    onFileSelected={editor.selectFile}
                    onRuleSelected={editor.selectRule}
                    overflowingElements={previewUpdate.metrics.overflowingElements}
                />
            </section>

            <section className="skin-editor-preview skin-editor-column">
                <div className="skin-preview-toolbar">
                    <div className="row skin-preview-group">
                        <label>Session</label>

                        <select
                            value={preview}
                            onChange={(e) => setPreview(e.target.value as "empty" | "running" | "completed")}
                        >
                            <option value="empty">Empty</option>
                            <option value="running">In Progress</option>
                            <option value="completed">Completed</option>
                        </select>
                    </div>

                    <div className="skin-preview-group">
                        <label>Comparison</label>

                        <div className="row button-row">
                            <button
                                onClick={() =>
                                    setComparison((current) => {
                                        const values = [
                                            CompareAgainst.Average,
                                            CompareAgainst.Best,
                                            CompareAgainst.SumOfBest,
                                        ];

                                        const index = values.indexOf(current);
                                        return values[(index - 1 + values.length) % values.length];
                                    })
                                }
                            >
                                ◀
                            </button>

                            <span>{comparison}</span>

                            <button
                                onClick={() =>
                                    setComparison((current) => {
                                        const values = [
                                            CompareAgainst.Average,
                                            CompareAgainst.Best,
                                            CompareAgainst.SumOfBest,
                                        ];

                                        const index = values.indexOf(current);
                                        return values[(index + 1) % values.length];
                                    })
                                }
                            >
                                ▶
                            </button>
                        </div>
                    </div>
                </div>
                <PreviewSplitter
                    skinCSS={model.styleSheet}
                    initialWidth={session.loaded_split_file?.window_width ?? 320}
                    initialHeight={session.loaded_split_file?.window_height ?? 580}
                    overrideCSS={overrideCSS}
                    sessionPayload={session}
                    configPayload={previewConfig}
                    elements={availableElements}
                    onPreviewUpdate={setPreviewUpdate}
                    selectedElement={selectedRuntime}
                    onSelect={editor.selectElement}
                    disableContextMenu
                    forceExpandAll
                    comparison={comparison}
                    onComparisonChange={setComparison}
                />
            </section>

            {hasPreviewOverflow && (
                <div className="skin-preview-warning">⚠ Elements extend outside the splitter preview.</div>
            )}

            <section className="skin-editor-files skin-editor-column">
                <SkinFiles
                    activeRule={editor.activeRule}
                    rules={editor.rules}
                    elements={allElements}
                    availableIds={availableIds}
                    selectedElement={editor.target.elementId}
                    files={editor.files}
                    selectedFile={editor.target.file}
                    mode={editor.target.mode}
                    selector={editor.target.selector}
                    revision={model.revision}
                    dirty={editor.dirty}
                    onElementSelected={editor.selectElement}
                    onSelectFile={editor.selectFile}
                    onSelectRule={editor.selectRule}
                    onCreateRule={editor.createRule}
                    onCreateFile={editor.createFile}
                    onChangeRule={editor.updateRule}
                    onChangeFile={editor.updateFile}
                    overflowingElements={previewUpdate.metrics.overflowingElements}
                />
            </section>

            <div className="skin-editor-actions">
                <button className="secondary" onClick={editor.reset} disabled={!editor.dirty}>
                    Reload
                </button>

                <button onClick={editor.save} disabled={!editor.dirty}>
                    Save
                </button>

                <button className="secondary" onClick={cancel}>
                    Cancel
                </button>
            </div>
        </div>
    );
}
