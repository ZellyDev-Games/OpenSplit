import type { CSSRuleEditor, SkinCSSRule } from "../../../models/skin/css";
import type { SkinModel } from "../../../models/skin/editor";
import type { SkinElement } from "../../../models/skin/element";
import { getElementRules } from "./elementTree/elementRules";
import { formatCSSClosingBraces, formatCSSParents, formatCSSRule } from "./elementTree/formatCSS";

interface Props {
    model: SkinModel;
    elements: SkinElement[];
    selectedElement: string | null;
    selectedFile: string | null;
    activeRule: CSSRuleEditor | null;
    availableIds: Set<string>;
    onFileSelected(file: string): Promise<void>;
    onRuleSelected(rule: CSSRuleEditor): Promise<void>;
}

export default function ElementTree({
    model,
    elements,
    selectedElement,
    selectedFile,
    activeRule,
    availableIds,
    onFileSelected,
    onRuleSelected,
}: Props) {
    const element = elements.find((item) => item.id === selectedElement);

    if (!element) {
        return <div className="element-tree">Select a preview element.</div>;
    }

    const available = availableIds.has(element.id);
    const rules = getElementRules(model, element);

    return (
        <div className="element-tree">
            {!available && (
                <div className="element-unavailable">✗ This element is not present in the current preview session.</div>
            )}

            {rules.length === 0 ? (
                <div>No CSS rules affect this element.</div>
            ) : (
                rules.map((rule) => (
                    <ElementRule
                        key={rule.id}
                        rule={rule}
                        selected={isRuleSelected(rule, activeRule, selectedFile)}
                        onFileSelected={onFileSelected}
                        onRuleSelected={onRuleSelected}
                    />
                ))
            )}
        </div>
    );
}

interface ElementRuleProps {
    rule: SkinCSSRule;
    selected: boolean;
    onFileSelected(file: string): Promise<void>;
    onRuleSelected(rule: CSSRuleEditor): Promise<void>;
}

function ElementRule({ rule, selected, onFileSelected, onRuleSelected }: ElementRuleProps) {
    async function handleClick(): Promise<void> {
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
            parentId: rule.parentId,
            parents: rule.parents ?? [],
            originalFile: rule.file,
            originalId: rule.id,
        });
    }

    return (
        <button type="button" className={selected ? "element-file selected" : "element-file"} onClick={handleClick}>
            <div className="file-name">{rule.file}</div>

            {rule.layer && <div className="layer">Layer: {rule.layer}</div>}

            <div className="line">Line {rule.line}</div>

            <pre>{formatRulePreview(rule)}</pre>
        </button>
    );
}

function isRuleSelected(rule: SkinCSSRule, activeRule: CSSRuleEditor | null, selectedFile: string | null): boolean {
    return activeRule?.id === rule.id || (!activeRule && selectedFile === rule.file);
}

function formatRulePreview(rule: SkinCSSRule): string {
    const parents = formatCSSParents(rule.parents);

    const closingBraces = rule.parents && rule.parents.length > 0 ? `\n${formatCSSClosingBraces(rule.parents)}` : "";

    return `${parents}${formatCSSRule(rule.selector, rule.body)}${closingBraces}`;
}
