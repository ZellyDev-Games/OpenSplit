export const CSS_LAYER_ORDER = ["reset", "components", "vars", "skins", "overrides"];

export function layerPriority(layer: string): number {
    const index = CSS_LAYER_ORDER.indexOf(layer);

    if (index >= 0) {
        return index;
    }

    // Unknown/runtime unlayered CSS has highest priority
    return CSS_LAYER_ORDER.length;
}

export function sortLayers(layers: string[]): string[] {
    return [...layers].sort((a, b) => layerPriority(a) - layerPriority(b));
}
