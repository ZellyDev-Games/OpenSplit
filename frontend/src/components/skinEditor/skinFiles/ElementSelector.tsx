import type { SkinModel } from "../../../models/skin/editor";
import type { SkinElement } from "../../../models/skin/element";
import { getSkinElementRules } from "../skinControls/elementTree/elementRules";

interface Props {
    model: SkinModel;

    elements: SkinElement[];

    overflowingIds: Set<string>;

    selectedElement: string | null;

    onElementSelected(id: string): void;
}

export default function ElementSelector({
    model,
    elements,
    overflowingIds,
    selectedElement,
    onElementSelected,
}: Props) {
    return (
        <>
            <h3>Element</h3>

            <select value={selectedElement ?? ""} onChange={(event) => onElementSelected(event.target.value)}>
                <option value="">Select Element</option>

                {elements.map((element) => {
                    const rules = getSkinElementRules(model, element);
                    const inSkin = rules.length > 0;
                    const overflowing = overflowingIds.has(element.id);
                    const icon = getElementStatusIcon(inSkin, overflowing);

                    return (
                        <option key={element.id} value={element.id}>
                            {icon}
                            {element.label}
                        </option>
                    );
                })}
            </select>
        </>
    );
}

function getElementStatusIcon(inSkin: boolean, overflowing: boolean): string {
    if (overflowing) {
        return "🔴 ";
    }

    return inSkin ? "✓ " : "✗ ";
}
