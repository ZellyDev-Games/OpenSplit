import { getBoxPadding, getMinimum } from "./css";
import type { MinimumSize } from "./types";

type ComponentDefinition = {
    selector: string;
    widthVariable: string;
    heightVariable: string;
};

/**
 * Component minimums are declared by the skin variables.
 *
 * The skin components use box-sizing: border-box, so the configured
 * minimum width/height already represents the complete component
 * minimum including its padding.
 *
 * Do not add component padding here.
 */
function getComponentMinimum(parent: HTMLElement, definition: ComponentDefinition): MinimumSize {
    const element = parent.querySelector<HTMLElement>(definition.selector);

    if (!element) {
        return {
            width: 0,
            height: 0,
        };
    }

    return {
        width: getMinimum(element, definition.widthVariable),

        height: getMinimum(element, definition.heightVariable),
    };
}

function getContainerMinimum(
    element: HTMLElement,
    definitions: ComponentDefinition[],
    direction: "row" | "column",
): MinimumSize {
    const sizes = definitions.map((definition) => getComponentMinimum(element, definition));

    const padding = getBoxPadding(element);

    if (direction === "row") {
        return {
            width: sizes.reduce((total, size) => total + size.width, 0) + padding.left + padding.right,

            height: Math.max(0, ...sizes.map((size) => size.height)) + padding.top + padding.bottom,
        };
    }

    return {
        width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,

        height: sizes.reduce((total, size) => total + size.height, 0) + padding.top + padding.bottom,
    };
}

const GAME_INFO_COMPONENTS: ComponentDefinition[] = [
    {
        selector: "#gameTitle",
        widthVariable: "--splitter-game-title-min-width",
        heightVariable: "--splitter-game-title-min-height",
    },
    {
        selector: "#gameCategory",
        widthVariable: "--splitter-game-category-min-width",
        heightVariable: "--splitter-game-category-min-height",
    },
    {
        selector: ".game-variable",
        widthVariable: "--splitter-game-variable-min-width",
        heightVariable: "--splitter-game-variable-min-height",
    },
    {
        selector: "#attempts",
        widthVariable: "--splitter-attempts-min-width",
        heightVariable: "--splitter-attempts-min-height",
    },
];

export function calculateGameInfoMinimumSize(element: HTMLElement): MinimumSize {
    return getContainerMinimum(element, GAME_INFO_COMPONENTS, "column");
}

const TIMER_COMPONENTS: ComponentDefinition[] = [
    {
        selector: "#time-sign",
        widthVariable: "--splitter-timer-sign-min-width",
        heightVariable: "--splitter-timer-sign-min-height",
    },
    {
        selector: "#time-hours",
        widthVariable: "--splitter-timer-hours-min-width",
        heightVariable: "--splitter-timer-hours-min-height",
    },
    {
        selector: "#time-sep-hm",
        widthVariable: "--splitter-timer-separator-hours-minutes-min-width",
        heightVariable: "--splitter-timer-separator-hours-minutes-min-height",
    },
    {
        selector: "#time-minutes",
        widthVariable: "--splitter-timer-minutes-min-width",
        heightVariable: "--splitter-timer-minutes-min-height",
    },
    {
        selector: "#time-sep-ms",
        widthVariable: "--splitter-timer-separator-minutes-seconds-min-width",
        heightVariable: "--splitter-timer-separator-minutes-seconds-min-height",
    },
    {
        selector: "#time-seconds",
        widthVariable: "--splitter-timer-seconds-min-width",
        heightVariable: "--splitter-timer-seconds-min-height",
    },
    {
        selector: "#time-sep-sc",
        widthVariable: "--splitter-timer-separator-seconds-centis-min-width",
        heightVariable: "--splitter-timer-separator-seconds-centis-min-height",
    },
    {
        selector: "#time-centis",
        widthVariable: "--splitter-timer-centis-min-width",
        heightVariable: "--splitter-timer-centis-min-height",
    },
];

export function calculateTimeContainerMinimumSize(element: HTMLElement): MinimumSize {
    return getContainerMinimum(element, TIMER_COMPONENTS, "row");
}

const WORLD_RECORD_PLAYER_COMPONENTS: ComponentDefinition[] = [
    {
        selector: "#world-record-label",
        widthVariable: "--splitter-world-record-label-min-width",
        heightVariable: "--splitter-world-record-label-min-height",
    },
    {
        selector: "#world-record-players",
        widthVariable: "--splitter-world-record-players-min-width",
        heightVariable: "--splitter-world-record-players-min-height",
    },
];

const WORLD_RECORD_RT_COMPONENTS: ComponentDefinition[] = [
    {
        selector: "#world-record-rt-label",
        widthVariable: "--splitter-world-record-rt-label-min-width",
        heightVariable: "--splitter-world-record-rt-label-min-height",
    },
    {
        selector: "#world-record-rt-time",
        widthVariable: "--splitter-world-record-rt-time-min-width",
        heightVariable: "--splitter-world-record-rt-time-min-height",
    },
    {
        selector: "#world-record-rt-centiseconds",
        widthVariable: "--splitter-world-record-rt-centiseconds-min-width",
        heightVariable: "--splitter-world-record-rt-centiseconds-min-height",
    },
];

const WORLD_RECORD_IGT_COMPONENTS: ComponentDefinition[] = [
    {
        selector: "#world-record-igt-label",
        widthVariable: "--splitter-world-record-igt-label-min-width",
        heightVariable: "--splitter-world-record-igt-label-min-height",
    },
    {
        selector: "#world-record-igt-time",
        widthVariable: "--splitter-world-record-igt-time-min-width",
        heightVariable: "--splitter-world-record-igt-time-min-height",
    },
    {
        selector: "#world-record-igt-centiseconds",
        widthVariable: "--splitter-world-record-igt-centiseconds-min-width",
        heightVariable: "--splitter-world-record-igt-centiseconds-min-height",
    },
];

function calculateWorldRecordSection(worldRecord: HTMLElement, definitions: ComponentDefinition[]): MinimumSize {
    return getContainerMinimum(worldRecord, definitions, "row");
}

export function calculateWorldRecordMinimumSize(element: HTMLElement): MinimumSize {
    const sections = [
        {
            selector: "#world-record-player",
            definitions: WORLD_RECORD_PLAYER_COMPONENTS,
        },
        {
            selector: "#world-record-real-time",
            definitions: WORLD_RECORD_RT_COMPONENTS,
        },
        {
            selector: "#world-record-in-game-time",
            definitions: WORLD_RECORD_IGT_COMPONENTS,
        },
    ];

    const sizes = sections
        .map(({ selector, definitions }) => {
            const section = element.querySelector<HTMLElement>(selector);

            if (!section) {
                return null;
            }

            return calculateWorldRecordSection(section, definitions);
        })
        .filter((size): size is MinimumSize => size !== null);

    const padding = getBoxPadding(element);

    return {
        width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,

        height: sizes.reduce((total, size) => total + size.height, 0) + padding.top + padding.bottom,
    };
}

export function calculateSplitterInfoMinimumSize(element: HTMLElement): MinimumSize {
    const comparison = element.querySelector<HTMLElement>(".comparison-mode");
    const timer = element.querySelector<HTMLElement>("#time-container");
    const worldRecord = element.querySelector<HTMLElement>("#world-record");

    const sizes: MinimumSize[] = [];

    if (comparison) {
        sizes.push({
            width: getMinimum(comparison, "--splitter-comparison-min-width"),

            height: getMinimum(comparison, "--splitter-comparison-min-height"),
        });
    }

    if (timer) {
        sizes.push(calculateTimeContainerMinimumSize(timer));
    }

    if (worldRecord) {
        sizes.push(calculateWorldRecordMinimumSize(worldRecord));
    }

    const padding = getBoxPadding(element);

    const layout = element.closest<HTMLElement>("#splitter")?.dataset.layout;

    if (layout === "horizontal") {
        return {
            width: sizes.reduce((total, size) => total + size.width, 0) + padding.left + padding.right,

            height: Math.max(0, ...sizes.map((size) => size.height)) + padding.top + padding.bottom,
        };
    }

    return {
        width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,

        height: sizes.reduce((total, size) => total + size.height, 0) + padding.top + padding.bottom,
    };
}
