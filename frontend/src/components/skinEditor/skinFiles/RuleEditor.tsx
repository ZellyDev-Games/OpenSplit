import { useEffect, useRef } from "react";

import type { CSSRuleEditor } from "../../../models/skin/css";
import type { SkinEditorTarget } from "../../../models/skin/editor";

interface Props {
    activeRule: CSSRuleEditor | null;

    rules: CSSRuleEditor[];

    mode: SkinEditorTarget["mode"];

    selector: string | null;

    body: string;

    layer: string | null;

    onBodyChange(value: string): void;

    onSelectRule(rule: CSSRuleEditor): void;

    onCreateRule(): Promise<void>;
}

export default function RuleEditor({
    activeRule,
    rules,
    mode,
    selector,
    body,
    layer,
    onBodyChange,
    onSelectRule,
    onCreateRule,
}: Props) {
    const textareaRef = useRef<HTMLTextAreaElement | null>(null);
    const previousRuleId = useRef<string | null>(activeRule?.id ?? null);

    /*
     * Selecting a different rule is an intentional document change.
     *
     * Load that rule's body directly into the textarea instead of
     * changing a React-controlled `value`.
     */
    useEffect(() => {
        const textarea = textareaRef.current;

        if (!textarea) {
            return;
        }

        const ruleId = activeRule?.id ?? null;

        if (previousRuleId.current === ruleId) {
            return;
        }

        previousRuleId.current = ruleId;

        textarea.value = body;
        textarea.scrollTop = 0;
        textarea.scrollLeft = 0;
    }, [activeRule?.id]);

    useEffect(() => {
        const textarea = textareaRef.current;

        if (!textarea || !activeRule) {
            return;
        }

        if (textarea.value === body) {
            return;
        }

        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const direction = textarea.selectionDirection;
        const scrollTop = textarea.scrollTop;
        const scrollLeft = textarea.scrollLeft;

        textarea.value = body;

        textarea.setSelectionRange(start, end, direction);
        textarea.scrollTop = scrollTop;
        textarea.scrollLeft = scrollLeft;
    }, [body]);
    /*
     * Synchronize external changes to the active rule.
     *
     * During normal typing, the backend sends the same value back.
     * Since the textarea already contains that value, this does nothing.
     *
     * If something else changes the rule, the textarea is updated.
     */
    useEffect(() => {
        const textarea = textareaRef.current;

        if (!textarea || !activeRule) {
            return;
        }

        if (textarea.value === body) {
            return;
        }

        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const direction = textarea.selectionDirection;
        const scrollTop = textarea.scrollTop;
        const scrollLeft = textarea.scrollLeft;

        textarea.value = body;

        textarea.setSelectionRange(start, end, direction);
        textarea.scrollTop = scrollTop;
        textarea.scrollLeft = scrollLeft;
    }, [body, activeRule]);

    function handleBodyChange(event: React.ChangeEvent<HTMLTextAreaElement>): void {
        onBodyChange(event.currentTarget.value);
    }

    function handleRuleChange(value: string): void {
        if (value === "__create__") {
            void onCreateRule();
            return;
        }

        const rule = rules.find((item) => item.id === value);

        if (rule) {
            onSelectRule(rule);
        }
    }

    return (
        <>
            <h3>Rule</h3>

            <h4>Selector</h4>

            <code>{selector}</code>

            {layer && (
                <>
                    <h4>Layer</h4>

                    <code>{layer}</code>
                </>
            )}

            {rules.length > 0 && (
                <>
                    <h4>Existing Rules</h4>

                    <select value={activeRule?.id ?? ""} onChange={(event) => handleRuleChange(event.target.value)}>
                        {rules.map((rule) => (
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

                    <textarea ref={textareaRef} defaultValue={body} onChange={handleBodyChange} />
                </div>
            )}
        </>
    );
}
