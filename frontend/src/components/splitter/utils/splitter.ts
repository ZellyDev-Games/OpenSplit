import { log } from "../../../utils/logger";
import { calculateGameInfoMinimumSize, calculateSplitterInfoMinimumSize } from "./components";
import { getBoxPadding } from "./css";
import { calculateSplitListMinimumSize } from "./splitList";
import type { MinimumSize } from "./types";

export type SplitterLayout = "vertical" | "horizontal";

function getChild(element: HTMLElement, id: string): HTMLElement | null {
    return (
        Array.from(element.children).find(
            (child): child is HTMLElement => child instanceof HTMLElement && child.id === id,
        ) ?? null
    );
}

export function calculateSplitterMinimumSize(element: HTMLElement): MinimumSize {
    const gameInfo = getChild(element, "gameInfo");
    const splitList = getChild(element, "splitList");
    const splitterInfo = getChild(element, "splitterInfo");

    const gameInfoMinimum = gameInfo ? calculateGameInfoMinimumSize(gameInfo) : { width: 0, height: 0 };
    log.debug(gameInfoMinimum);

    const splitListMinimum = splitList ? calculateSplitListMinimumSize(splitList) : { width: 0, height: 0 };
    log.debug(splitListMinimum);

    const splitterInfoMinimum = splitterInfo ? calculateSplitterInfoMinimumSize(splitterInfo) : { width: 0, height: 0 };
    log.debug(splitterInfoMinimum);

    const padding = getBoxPadding(element);

    const layout = element.dataset.layout as SplitterLayout | undefined;

    if (layout === "horizontal") {
        return {
            width:
                gameInfoMinimum.width +
                splitListMinimum.width +
                splitterInfoMinimum.width +
                padding.left +
                padding.right,

            height:
                Math.max(gameInfoMinimum.height, splitListMinimum.height, splitterInfoMinimum.height) +
                padding.top +
                padding.bottom,
        };
    }

    /*
     * Vertical:
     *
     * gameInfo
     * splitList
     * splitterInfo
     */
    return {
        width:
            Math.max(gameInfoMinimum.width, splitListMinimum.width, splitterInfoMinimum.width) +
            padding.left +
            padding.right,

        height:
            gameInfoMinimum.height +
            splitListMinimum.height +
            splitterInfoMinimum.height +
            padding.top +
            padding.bottom,
    };
}
