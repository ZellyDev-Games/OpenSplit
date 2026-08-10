import type { CSSRuleEditor } from "../../models/skin/css";
import type { SkinModel } from "../../models/skin/editor";
import type { SkinElement } from "../../models/skin/element";
import ElementTree from "./skinControls/ElementTree";
import PreviewElementSelect from "./skinControls/PreviewElementSelect";
import { getSkinControlsInfo } from "./skinControls/skinControlsInfo";

interface Props {
    model: SkinModel;

    elements: SkinElement[];

    selectedElement: string | null;

    selectedFile: string | null;

    activeRule: CSSRuleEditor | null;

    onElementSelected(id: string): void;

    onFileSelected(file: string): Promise<void>;

    onRuleSelected(rule: CSSRuleEditor): Promise<void>;

    overflowingIds: Set<string>;

    availableIds: Set<string>;
}

export default function SkinControls({
    model,
    elements,
    selectedElement,
    selectedFile,
    activeRule,
    onElementSelected,
    onFileSelected,
    onRuleSelected,
    overflowingIds,
    availableIds,
}: Props) {
    const info = getSkinControlsInfo(model, elements, selectedElement);

    return (
        <div className="skin-controls">
            <h3>Editing Skin</h3>

            <div>{model.name}</div>

            <hr />

            <h3>Preview Element</h3>

            <PreviewElementSelect
                model={model}
                elements={elements}
                selectedElement={selectedElement}
                overflowingIds={overflowingIds}
                onElementSelected={onElementSelected}
            />

            <div className="skin-selector-info">
                <label>Selector</label>

                <code>{info.selector}</code>

                <label>Layers</label>

                <code>{info.layers}</code>

                <label>Rules</label>

                <code>{info.ruleCount}</code>
            </div>

            <hr />

            <ElementTree
                model={model}
                elements={elements}
                selectedElement={selectedElement}
                selectedFile={selectedFile}
                activeRule={activeRule}
                availableIds={availableIds}
                onFileSelected={onFileSelected}
                onRuleSelected={onRuleSelected}
            />
        </div>
    );
}
