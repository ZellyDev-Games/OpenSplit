import type { RuntimeElement } from "../../../models/skin/element";
import type { PreviewMetrics } from "../../../models/skin/preview";
import type { Highlight } from "../useElementHighlight";

export function createPreviewHighlights(
    runtimeElements: RuntimeElement[],
    metrics: PreviewMetrics,
    selectedElement: RuntimeElement | null,
): Highlight[] {
    return runtimeElements.flatMap<Highlight>((element) => {
        const selected = selectedElement?.id === element.id;

        const overflow = metrics.overflowingIds.has(element.id);

        if (selected && overflow) {
            return [
                {
                    element: element.element,
                    type: "selected-overflow",
                },
            ];
        }

        if (selected) {
            return [
                {
                    element: element.element,
                    type: "selected",
                },
            ];
        }

        if (overflow) {
            return [
                {
                    element: element.element,
                    type: "overflow",
                },
            ];
        }

        return [];
    });
}
