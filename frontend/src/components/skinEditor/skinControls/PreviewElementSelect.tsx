import type { SkinModel } from "../../../models/skin/editor";
import type { SkinElement } from "../../../models/skin/element";
import { log } from "../../../utils/logger";
import { getSkinElementRules } from "./elementTree/elementRules";

interface Props {
    model: SkinModel;

    elements: SkinElement[];

    overflowingIds: Set<string>;

    selectedElement: string | null;

    onElementSelected(id: string): void;
}

export default function PreviewElementSelect({
    model,
    elements,
    overflowingIds,
    selectedElement,
    onElementSelected,
}: Props) {
    return (
        <select value={selectedElement ?? ""} onChange={(event) => onElementSelected(event.target.value)}>
            <option value="">Select element…</option>

            {elements.map((element) => {
                const rules = getSkinElementRules(model, element);
                const inSkin = rules.length > 0;
                const overflowing = overflowingIds.has(element.id);
                const icon = getElementStatusIcon(inSkin, overflowing);

                log.debug("[SkinControls] Preview element dropdown", {
                    id: element.id,
                    label: element.label,
                    selector: element.selector,
                    inSkin,
                    ruleCount: rules.length,
                    overflowing,
                    icon: icon.trim(),
                });

                return (
                    <option key={element.id} value={element.id} disabled={!inSkin}>
                        {icon}
                        {element.label}
                    </option>
                );
            })}
        </select>
    );
}

function getElementStatusIcon(inSkin: boolean, overflowing: boolean): string {
    if (overflowing) {
        return "🔴 ";
    }

    return inSkin ? "✓ " : "✗ ";
}
