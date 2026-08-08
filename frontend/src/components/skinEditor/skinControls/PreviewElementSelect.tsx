import type { SkinElement } from "../../../models/skin/element";

interface Props {
    elements: SkinElement[];

    availableIds: Set<string>;

    selectedElement: string | null;

    overflowingIds: Set<string>;

    onElementSelected(id: string): void;
}

export default function PreviewElementSelect({
    elements,
    availableIds,
    selectedElement,
    overflowingIds,
    onElementSelected,
}: Props) {
    return (
        <select value={selectedElement ?? ""} onChange={(event) => onElementSelected(event.target.value)}>
            <option value="">Select element…</option>

            {elements.map((element) => {
                const available = availableIds.has(element.id);
                const overflowing = overflowingIds.has(element.id);

                return (
                    <option key={element.id} value={element.id} disabled={!available}>
                        {getElementStatusIcon(available, overflowing)}
                        {element.label}
                    </option>
                );
            })}
        </select>
    );
}

function getElementStatusIcon(available: boolean, overflowing: boolean): string {
    if (overflowing) {
        return "🔴 ";
    }

    return available ? "✓ " : "✗ ";
}
