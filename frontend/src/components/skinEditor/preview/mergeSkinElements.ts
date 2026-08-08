import type { RuntimeElement, SkinElement } from "../../../models/skin/element";

export function mergeSkinElements(
    preview: SkinElement[],
    backend: SkinElement[],
    runtime: RuntimeElement[],
): SkinElement[] {
    const backendMap = new Map(backend.map((element) => [element.id, element]));

    const runtimeMap = new Map(runtime.map((element) => [element.id, element]));

    return preview.map((element) => ({
        ...element,

        // backend adds editable metadata
        ...(backendMap.get(element.id) ?? {}),

        // runtime adds actual DOM node
        ...(runtimeMap.get(element.id) ?? {}),
    }));
}
