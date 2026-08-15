import { useCallback, useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { useSkinEditor } from "../../hooks/skinEditor/useSkinEditor";
import { useSkinEditorPreview } from "../../hooks/skinEditor/useSkinEditorPreview";
import { SplitterLayout } from "../../hooks/splitter/useSplitterMenu";
import { Command } from "../../models/command";
import type { SkinModel } from "../../models/skin/editor";
import PreviewSplitter from "./PreviewSplitter";
import SkinControls from "./SkinControls";
import SkinEditorToolbar from "./SkinEditorToolbar";
import SkinFiles from "./SkinFiles";

interface Props {
    model: SkinModel;
}

const PREVIEW_SIZES: Record<SplitterLayout, { width: number; height: number }> = {
    vertical: {
        width: 350,
        height: 550,
    },
    horizontal: {
        width: 800,
        height: 350,
    },
};

export default function SkinEditor({ model }: Props) {
    const editor = useSkinEditor(model);
    const preview = useSkinEditorPreview(model, editor);

    const [layout, setLayout] = useState<SplitterLayout>("vertical");

    async function setDefaultLayout() {
        await Dispatch(
            Command.SKIN_SET_DEFAULT_LAYOUT,
            JSON.stringify({
                layout,
            }),
        );
    }

    const cancel = useCallback(async () => {
        await Dispatch(Command.CANCEL, null);
    }, []);

    const { width: previewWidth, height: previewHeight } = PREVIEW_SIZES[layout];

    return (
        <div className="skin-editor">
            <section className="skin-editor-left panel skin-editor-column">
                <SkinControls
                    model={preview.combinedModel}
                    elements={preview.elements}
                    selectedElement={editor.target.elementId}
                    selectedFile={editor.target.file}
                    activeRule={editor.activeRule}
                    onElementSelected={editor.selectElement}
                    onFileSelected={editor.selectFile}
                    onRuleSelected={editor.selectRule}
                    overflowingIds={preview.previewUpdate.metrics.overflowingIds}
                    availableIds={preview.availableIds}
                />
            </section>

            <section className="skin-editor-files skin-editor-column">
                <SkinFiles
                    model={preview.combinedModel}
                    activeRule={editor.activeRule}
                    rules={editor.rules}
                    elements={preview.elements}
                    selectedElement={editor.target.elementId}
                    files={editor.files}
                    selectedFile={editor.target.file}
                    mode={editor.target.mode}
                    selector={editor.target.selector}
                    layer={editor.activeRule?.layer ?? null}
                    onElementSelected={editor.selectElement}
                    onSelectFile={editor.selectFile}
                    onSelectRule={editor.selectRule}
                    onCreateRule={editor.createRule}
                    onCreateFile={editor.createFile}
                    onChangeRule={editor.updateRule}
                    onChangeFile={editor.updateFile}
                    overflowingIds={preview.previewUpdate.metrics.overflowingIds}
                />
            </section>

            <section className="skin-editor-toolbar">
                <SkinEditorToolbar
                    mode={preview.mode}
                    comparison={preview.comparison}
                    layout={layout}
                    onModeChange={preview.setMode}
                    onComparisonChange={preview.setComparison}
                    onLayoutChange={setLayout}
                    onSetDefaultLayout={setDefaultLayout}
                />
            </section>

            <div className="skin-editor-preview-area">
                <section className="skin-editor-preview skin-editor-column">
                    <PreviewSplitter
                        skinCSS={model.styleSheet}
                        initialWidth={previewWidth}
                        initialHeight={previewHeight}
                        overrideCSS={preview.overrideCSS}
                        sessionPayload={preview.session}
                        configPayload={preview.config}
                        elements={preview.elements}
                        onPreviewUpdate={preview.setPreviewUpdate}
                        selectedElement={preview.selectedRuntime}
                        onSelect={editor.selectElement}
                        disableContextMenu
                        forceExpandAll
                        comparison={preview.comparison}
                        onComparisonChange={preview.setComparison}
                        layout={layout}
                        onLayoutChange={setLayout}
                    />
                </section>

                {preview.hasPreviewOverflow && (
                    <div className="skin-preview-warning">⚠ Elements extend outside the splitter preview.</div>
                )}
            </div>

            <div className="skin-editor-actions">
                {editor.dirty && <div className="unsaved-changes">Unsaved changes</div>}

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
