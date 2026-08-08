import type { SkinElement } from "../../../models/skin/element";

interface Props {
    elements: SkinElement[];

    availableIds: Set<string>;

    overflowingIds: Set<string>;

    selectedElement: string | null;

    onElementSelected(id: string): void;
}

export default function ElementSelector({
    elements,
    availableIds,
    overflowingIds,
    selectedElement,
    onElementSelected,
}: Props) {
    const overflowing = new Set(overflowingIds);

    return (
        <>
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
        </>
    );
}
