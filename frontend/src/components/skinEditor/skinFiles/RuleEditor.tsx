import type { CSSRuleEditor } from "../../../models/skin/css";
import type { SkinEditorTarget } from "../../../models/skin/editor";

interface Props {
    activeRule: CSSRuleEditor | null;

    rules: CSSRuleEditor[];

    mode: SkinEditorTarget["mode"];

    selector: string | null;

    body: string;

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
    onBodyChange,
    onSelectRule,
    onCreateRule,
}: Props) {
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

                    <textarea value={body} onChange={(event) => onBodyChange(event.target.value)} />
                </div>
            )}
        </>
    );
}
