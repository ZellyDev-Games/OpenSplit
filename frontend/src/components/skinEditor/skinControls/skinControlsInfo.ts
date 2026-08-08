import type { SkinModel } from "../../../models/skin/editor";
import type { SkinElement } from "../../../models/skin/element";
import { sortLayers } from "./utils/layerOrder";
import { selectorMatches } from "./utils/selectorMatch";

export interface SkinControlsInfo {
    selector: string;
    layers: string;
    ruleCount: number;
}

export function getSkinControlsInfo(
    model: SkinModel,
    elements: SkinElement[],
    selectedElement: string | null,
): SkinControlsInfo {
    const currentElement = elements.find((element) => element.id === selectedElement);

    if (!currentElement) {
        return {
            selector: "(none)",
            layers: "(none)",
            ruleCount: 0,
        };
    }

    const currentRules = model.rules.filter((rule) => selectorMatches(rule.selector, currentElement.selector));

    const layers = sortLayers([...new Set(currentRules.map((rule) => rule.layer).filter((layer) => layer.length > 0))]);

    return {
        selector: currentElement.selector,
        layers: layers.length > 0 ? layers.join(", ") : "(none)",
        ruleCount: currentRules.length,
    };
}
