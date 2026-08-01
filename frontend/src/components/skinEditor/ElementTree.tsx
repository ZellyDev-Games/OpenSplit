import { CSSRuleEditor } from "../../models/cssProperty";
import SkinModel, { CSSAtRule, SkinElement } from "../../models/skinModel";
import { formatCSSRule } from "./formatCSS";
import { selectorMatches } from "./selectorMatch";

interface Props {
    model: SkinModel;

    elements: SkinElement[];

    availableIds: Set<string>;

    selectedElement: string | null;

    selectedFile: string | null;

    activeRule: CSSRuleEditor | null;

    onFileSelected(file: string): Promise<void>;

    onRuleSelected(rule: CSSRuleEditor): Promise<void>;
}

function formatParents(parents: CSSAtRule[] = []): string {
    if (parents.length === 0) {
        return "";
    }

    return (
        parents
            .map((parent) => {
                switch (parent.type) {
                    case "layer":
                        return `@layer ${parent.name} {`;

                    case "media":
                        return `@media ${parent.params ?? ""} {`;

                    case "supports":
                        return `@supports ${parent.params ?? ""} {`;

                    case "font-face":
                        return "@font-face {";

                    case "import":
                        return `@import ${parent.params ?? parent.name};`;

                    default:
                        return `@${parent.name}${parent.params ? ` ${parent.params}` : ""} {`;
                }
            })
            .join("\n") + "\n"
    );
}

function closingBraces(parents: CSSAtRule[] = []): string {
    return parents
        .filter((p) => p.type !== "import")
        .map(() => "}")
        .reverse()
        .join("\n");
}

export default function ElementTree({
    model,
    elements,
    availableIds,
    selectedElement,
    selectedFile,
    activeRule,
    onFileSelected,
    onRuleSelected,
}: Props) {
    const element = elements.find((item) => item.id === selectedElement);

    if (!element) {
        return <div className="element-tree">Select a preview element.</div>;
    }

    const available = availableIds.has(element.id);

    const rules = model.rules
        .filter((rule) => selectorMatches(rule.selector, element.selector))
        .sort((a, b) => {
            if (a.file !== b.file) {
                return a.file.localeCompare(b.file);
            }

            if (a.line !== b.line) {
                return a.line - b.line;
            }

            return a.order - b.order;
        });

    if (rules.length === 0) {
        return <div className="element-tree">No CSS rules affect this element.</div>;
    }

    return (
        <div className="element-tree">
            {!available && (
                <div className="element-unavailable">✗ This element is not present in the current preview session.</div>
            )}

            {rules.map((rule) => {
                const selected = activeRule?.id === rule.id || (!activeRule && selectedFile === rule.file);

                return (
                    <button
                        key={rule.id}
                        type="button"
                        className={selected ? "element-file selected" : "element-file"}
                        onClick={async () => {
                            if (rule.file === "runtime") {
                                return;
                            }

                            await onFileSelected(rule.file);

                            await onRuleSelected({
                                id: rule.id,
                                file: rule.file,
                                selector: rule.selector,
                                layer: rule.layer,
                                body: rule.body,
                                parents: rule.parents ?? [],
                                originalFile: rule.file,
                                originalId: rule.id,
                            });
                        }}
                    >
                        <div className="file-name">{rule.file}</div>

                        {rule.layer && <div className="layer">Layer: {rule.layer}</div>}

                        <div className="line">Line {rule.line}</div>

                        <pre>
                            {formatParents(rule.parents)}
                            {formatCSSRule(rule.selector, rule.body)}
                            {rule.parents && rule.parents.length > 0 && "\n" + closingBraces(rule.parents)}
                        </pre>
                    </button>
                );
            })}
        </div>
    );
}
