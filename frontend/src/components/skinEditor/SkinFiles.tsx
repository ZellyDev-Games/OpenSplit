import { useEffect, useRef, useState } from "react";

import type { CSSFile, CSSRuleEditor } from "../../models/skin/css";
import type { SkinEditorTarget } from "../../models/skin/editor";
import type { SkinElement } from "../../models/skin/element";
import ElementSelector from "./skinFiles/ElementSelector";
import FileSelector from "./skinFiles/FileSelector";
import { isCSSFile, isImageFile } from "./skinFiles/fileType";
import ImagePreview from "./skinFiles/ImagePreview";
import RuleEditor from "./skinFiles/RuleEditor";
import TextFileEditor from "./skinFiles/TextFileEditor";

interface Props {
    activeRule: CSSRuleEditor | null;

    rules: CSSRuleEditor[];

    availableIds: Set<string>;

    elements: SkinElement[];

    selectedElement: string | null;

    files: CSSFile[];

    selectedFile: string | null;

    mode: SkinEditorTarget["mode"];

    selector: string | null;

    layer: string | null;

    onElementSelected(id: string): void;

    onSelectFile(path: string): Promise<void>;

    onSelectRule(rule: CSSRuleEditor): void;

    onCreateRule(): Promise<void>;

    onCreateFile(name: string): Promise<void>;

    onChangeRule(rule: CSSRuleEditor): Promise<void>;

    onChangeFile(file: string, contents: string): Promise<void>;

    revision: number;

    dirty: boolean;

    overflowingIds: Set<string>;
}

export default function SkinFiles({
    activeRule,
    rules,
    availableIds,
    elements,
    selectedElement,
    files,
    selectedFile,
    mode,
    selector,
    layer,
    revision,
    dirty,
    onElementSelected,
    onSelectFile,
    onSelectRule,
    onCreateRule,
    onCreateFile,
    onChangeRule,
    onChangeFile,
    overflowingIds,
}: Props) {
    const [body, setBody] = useState(activeRule?.body ?? "");

    const editingRule = useRef<string | null>(null);
    const lastRevision = useRef<number>(revision);

    useEffect(() => {
        if (!activeRule) {
            editingRule.current = null;
            setBody("");
            return;
        }

        const ruleChanged = editingRule.current !== activeRule.id;
        const reloaded = lastRevision.current !== revision;

        if (ruleChanged || reloaded) {
            editingRule.current = activeRule.id;
            lastRevision.current = revision;
            setBody(activeRule.body);
        }
    }, [activeRule, revision]);

    const selectedFileData = files.find((file) => file.path === selectedFile);

    const cssRules = rules.filter((rule) => rule.file === selectedFile);

    const imageFile = selectedFileData && isImageFile(selectedFileData) ? selectedFileData : null;

    const cssFile = selectedFileData && isCSSFile(selectedFileData);

    return (
        <div className="panel skin-files">
            <ElementSelector
                elements={elements}
                availableIds={availableIds}
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

            {cssFile && (
                <RuleEditor
                    activeRule={activeRule}
                    rules={cssRules}
                    mode={mode}
                    selector={selector}
                    layer={layer}
                    body={body}
                    onBodyChange={(value) => {
                        setBody(value);

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

            {selectedFileData?.text && !cssFile && (
                <TextFileEditor file={selectedFileData} onChangeFile={onChangeFile} />
            )}

            {dirty && <div className="unsaved-changes">Unsaved changes</div>}
        </div>
    );
}
