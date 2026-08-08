import { CSSRuleEditor } from "../../models/skin/css";
import { SkinModel } from "../../models/skin/editor";
import { SkinElement } from "../../models/skin/element";
import ElementTree from "./skinControls/ElementTree";
import { sortLayers } from "./skinControls/layerOrder";
import { selectorMatches } from "./utils/selectorMatch";

interface Props {
    model: SkinModel;

    elements: SkinElement[];

    availableIds: Set<string>;

    selectedElement: string | null;

    selectedFile: string | null;

    activeRule: CSSRuleEditor | null;

    onElementSelected(id: string): void;

    onFileSelected(file: string): Promise<void>;

    onRuleSelected(rule: CSSRuleEditor): Promise<void>;

    overflowingIds: Set<string>;
}

export default function SkinControls({
    model,
    elements,
    availableIds,
    selectedElement,
    selectedFile,
    activeRule,
    onElementSelected,
    onFileSelected,
    onRuleSelected,
    overflowingIds,
}: Props) {
    const currentElement = elements.find((element) => element.id === selectedElement) ?? null;

    const overflowing = new Set(overflowingIds);

    const currentRules = currentElement
        ? model.rules.filter((rule) => selectorMatches(rule.selector, currentElement.selector))
        : [];

    const layers = sortLayers([...new Set(currentRules.map((rule) => rule.layer).filter((layer) => layer.length > 0))]);

    return (
        <div className="skin-controls">
            <h3>Editing Skin</h3>

            <div>{model.name}</div>

            <hr />

            <h3>Preview Element</h3>

            <select value={selectedElement ?? ""} onChange={(event) => onElementSelected(event.target.value)}>
                <option value="">Select element…</option>

                {elements.map((element) => {
                    const available = availableIds.has(element.id);

                    return (
                        <option key={element.id} value={element.id} disabled={!available}>
                            {overflowing.has(element.id) ? "🔴 " : available ? "✓ " : "✗ "}
                            {element.label}
                        </option>
                    );
                })}
            </select>

            <div className="skin-selector-info">
                <label>Selector</label>

                <code>{currentElement?.selector ?? "(none)"}</code>

                <label>Layers</label>

                <code>{layers.length ? layers.join(", ") : "(none)"}</code>

                <label>Rules</label>

                <code>{currentRules.length}</code>
            </div>

            <hr />

            <ElementTree
                model={model}
                elements={elements}
                availableIds={availableIds}
                selectedElement={selectedElement}
                selectedFile={selectedFile}
                activeRule={activeRule}
                onFileSelected={onFileSelected}
                onRuleSelected={onRuleSelected}
            />
        </div>
    );
}
