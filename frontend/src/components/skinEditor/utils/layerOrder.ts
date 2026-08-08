export const SKIN_LAYER_ORDER = ["reset", "vars", "components", "skins", "overrides"] as const;

const layerIndex = new Map<string, number>(SKIN_LAYER_ORDER.map((layer, index) => [layer, index]));

export function layerPriority(layer?: string): number {
    if (!layer) {
        // Unlayered CSS has highest cascade priority.
        return Number.MAX_SAFE_INTEGER;
    }

    return layerIndex.get(layer) ?? Number.MAX_SAFE_INTEGER - 1;
}

export function sortLayers(layers: string[]): string[] {
    return [...layers].sort((a, b) => layerPriority(a) - layerPriority(b));
}
