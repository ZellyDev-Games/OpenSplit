import type { CSSFile, CSSRuleEditor } from "../../models/skin/css";
import type { SkinEditorTarget, SkinModel } from "../../models/skin/editor";
import type { SkinElement } from "../../models/skin/element";
import ElementSelector from "./skinFiles/ElementSelector";
import FileSelector from "./skinFiles/FileSelector";
import { isCSSFile, isImageFile } from "./skinFiles/fileType";
import ImagePreview from "./skinFiles/ImagePreview";
import RuleEditor from "./skinFiles/RuleEditor";
import TextFileEditor from "./skinFiles/TextFileEditor";

interface Props {
    model: SkinModel;

    activeRule: CSSRuleEditor | null;

    rules: CSSRuleEditor[];

    elements: SkinElement[];

    selectedElement: string | null;

    files: CSSFile[];

    selectedFile: string | null;

    mode: SkinEditorTarget["mode"];

    selector: string | null;

    layer: string | null;

    onElementSelected(id: string): void;

    onSelectFile(path: string): Promise<void>;

    onCreateRule(): Promise<void>;

    onSelectRule(rule: CSSRuleEditor): void;

    onCreateFile(name: string): Promise<void>;

    onChangeRule(rule: CSSRuleEditor): Promise<void>;

    onChangeFile(file: string, contents: string): Promise<void>;

    overflowingIds: Set<string>;
}

export default function SkinFiles({
    model,
    activeRule,
    rules,
    elements,
    selectedElement,
    files,
    selectedFile,
    mode,
    selector,
    layer,
    onElementSelected,
    onSelectFile,
    onCreateFile,
    onSelectRule,
    onCreateRule,
    onChangeRule,
    onChangeFile,
    overflowingIds,
}: Props) {
    const selectedFileData = files.find((file) => file.path === selectedFile);

    const cssRules = rules.filter((rule) => rule.file === selectedFile);

    const imageFile = selectedFileData && isImageFile(selectedFileData) ? selectedFileData : null;

    const cssFile = selectedFileData && isCSSFile(selectedFileData);

    /*
     * The element selector uses an empty string for "no element selected".
     * Normalize that here so both null and "" represent the unselected state.
     */
    const hasSelectedElement = selectedElement !== null && selectedElement !== "";

    /*
     * Raw file editing is the default when no element is selected.
     *
     * Once an element is selected, an existing rule, rule mode, or an
     * available layer can switch the CSS file into the structured rule
     * editor.
     */
    const fileHasLayer = cssRules.some((rule) => rule.layer !== null && rule.layer !== "");

    const showRuleEditor =
        cssFile && hasSelectedElement && (mode !== "file" || activeRule !== null || layer !== null || fileHasLayer);

    return (
        <div className="panel skin-files">
            <ElementSelector
                model={model}
                elements={elements}
                overflowingIds={overflowingIds}
                selectedElement={selectedElement}
                onElementSelected={onElementSelected}
            />

            <FileSelector
                files={files}
                selectedFile={selectedFile}
                onSelectFile={onSelectFile}
                onCreateFile={onCreateFile}
            />

            {imageFile && <ImagePreview file={imageFile} />}

            {selectedFileData?.text && !showRuleEditor && (
                <TextFileEditor file={selectedFileData} onChangeFile={onChangeFile} />
            )}

            {showRuleEditor && (
                <RuleEditor
                    activeRule={activeRule}
                    rules={cssRules}
                    mode={mode}
                    selector={selector}
                    layer={layer}
                    body={activeRule?.body ?? ""}
                    onBodyChange={(value) => {
                        if (!activeRule) {
                            return;
                        }

                        void onChangeRule({
                            ...activeRule,
                            body: value,
                        });
                    }}
                    onSelectRule={onSelectRule}
                    onCreateRule={onCreateRule}
                />
            )}
        </div>
    );
}
