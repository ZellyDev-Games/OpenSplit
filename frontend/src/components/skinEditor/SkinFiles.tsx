import { useEffect, useRef, useState } from "react";

import { CSSFile, CSSRuleEditor } from "../../models/skin/css";
import { SkinEditorTarget } from "../../models/skin/editor";
import { SkinElement } from "../../models/skin/element";

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

function TextFileEditor({
    file,
    onChangeFile,
}: {
    file: CSSFile;
    onChangeFile(file: string, contents: string): Promise<void>;
}) {
    const [contents, setContents] = useState(file.contents);

    useEffect(() => {
        setContents(file.contents);
    }, [file.path, file.contents]);

    return (
        <div className="panel text-file-editor">
            <h3>Text</h3>

            <textarea
                value={contents}
                onChange={(event) => {
                    const value = event.target.value;

                    setContents(value);

                    void onChangeFile(file.path, value);
                }}
            />
        </div>
    );
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

    const overflowing = new Set(overflowingIds);

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

    const selected = files.find((file) => file.path === selectedFile);

    const cssRules = rules.filter((rule) => rule.file === selectedFile);

    const isImage =
        selected &&
        !selected.text &&
        (selected.type.startsWith("image") || /\.(png|jpg|jpeg|gif|svg|webp)$/i.test(selected.path));

    const isCSS = selected?.type === "css" || selected?.path.endsWith(".css");

    return (
        <div className="panel skin-files">
            <h3>Element</h3>

            <select value={selectedElement ?? ""} onChange={(event) => onElementSelected(event.target.value)}>
                <option value="">Select Element</option>

                {elements.map((element) => {
                    const available = availableIds.has(element.id);

                    return (
                        <option key={element.id} value={element.id}>
                            {overflowing.has(element.id) ? "🔴 " : available ? "✓ " : "✗ "}
                            {element.label}
                        </option>
                    );
                })}
            </select>

            <h3>CSS File</h3>

            <select
                value={selectedFile ?? ""}
                onChange={(event) => {
                    if (event.target.value === "__new__") {
                        const name = prompt("CSS filename");

                        if (name) {
                            void onCreateFile(name.endsWith(".css") ? name : `${name}.css`);
                        }

                        return;
                    }

                    onSelectFile(event.target.value);
                }}
            >
                <option value="">Select file</option>

                {files.map((file) => (
                    <option key={file.path} value={file.path}>
                        {file.name}
                    </option>
                ))}

                <option value="__new__">Create new CSS file…</option>
            </select>

            {isImage && (
                <>
                    <h3>Preview</h3>

                    <div className="skin-image-preview">
                        <img src={selected.url} alt={selected.name} />
                    </div>
                </>
            )}

            {isCSS && (
                <>
                    <h3>Rule</h3>

                    <h4>Selector</h4>

                    <code>{selector}</code>

                    {cssRules.length > 0 && (
                        <>
                            <h4>Existing Rules</h4>

                            <select
                                value={activeRule?.id ?? ""}
                                onChange={(event) => {
                                    if (event.target.value === "__create__") {
                                        onCreateRule();

                                        return;
                                    }

                                    const rule = cssRules.find((item) => item.id === event.target.value);

                                    if (rule) {
                                        onSelectRule(rule);
                                    }
                                }}
                            >
                                {cssRules.map((rule) => (
                                    <option key={rule.id} value={rule.id}>
                                        {rule.selector}

                                        {rule.layer ? ` (${rule.layer})` : ""}
                                    </option>
                                ))}

                                <option value="__create__">+ Create new rule</option>
                            </select>
                        </>
                    )}

                    {mode === "create" && (
                        <button
                            type="button"
                            onClick={() => {
                                void onCreateRule();
                            }}
                        >
                            Create Rule
                        </button>
                    )}

                    {activeRule && (
                        <div className="panel rule-editor">
                            <h4>CSS</h4>

                            <textarea
                                value={body}
                                onChange={(e) => {
                                    const value = e.target.value;

                                    setBody(value);

                                    if (!activeRule) {
                                        return;
                                    }

                                    void onChangeRule({
                                        ...activeRule,
                                        body: value,
                                    });
                                }}
                            />
                        </div>
                    )}
                </>
            )}

            {selected?.text && !isCSS && <TextFileEditor file={selected} onChangeFile={onChangeFile} />}
            {dirty && <div className="unsaved-changes">Unsaved changes</div>}
        </div>
    );
}
