import { getBoxBorder, getBoxPadding, getContentMinimumSize, getEffectiveMinimumSize, getGap } from "./css";
import type { MinimumSize } from "./types";

type ComponentDefinition = {
    selector: string;
    widthVariable: string;
    heightVariable: string;
    contentAware?: {
        width?: boolean;
        height?: boolean;
    };
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

    return getEffectiveMinimumSize(
        element,
        definition.widthVariable,
        definition.heightVariable,
        definition.contentAware,
    );
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
        contentAware: {
            // width: true,
            height: true,
        },
    },
    {
        selector: "#gameCategory",
        widthVariable: "--splitter-game-category-min-width",
        heightVariable: "--splitter-game-category-min-height",
        contentAware: {
            // width: true,
            height: true,
        },
    },
    {
        selector: ".game-variable",
        widthVariable: "--splitter-game-variable-min-width",
        heightVariable: "--splitter-game-variable-min-height",
        contentAware: {
            // width: true,
            height: true,
        },
    },
    {
        selector: "#attempts",
        widthVariable: "--splitter-attempts-min-width",
        heightVariable: "--splitter-attempts-min-height",
    },
];

export function calculateGameInfoMinimumSize(element: HTMLElement): MinimumSize {
    const layout = element.closest<HTMLElement>("#splitter")?.dataset.layout;
    const sizes = GAME_INFO_COMPONENTS.flatMap((definition) => {
        if (definition.selector !== ".game-variable") {
            return [getComponentMinimum(element, definition)];
        }

        return Array.from(element.querySelectorAll<HTMLElement>(definition.selector)).map((variable) =>
            getEffectiveMinimumSize(
                variable,
                definition.widthVariable,
                definition.heightVariable,
                definition.contentAware,
            ),
        );
    });
    const padding = getBoxPadding(element);

    if (layout === "horizontal") {
        return {
            width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,
            height: Math.max(0, ...sizes.map((size) => size.height)) + padding.top + padding.bottom,
        };
    }

    const border = getBoxBorder(element);
    const outOfFlowDescendants = Array.from(element.querySelectorAll<HTMLElement>("*"))
        .filter((child) => {
            const position = getComputedStyle(child).position;
            return position === "absolute" || position === "fixed";
        })
        .map((child) => ({
            child,
            display: child.style.getPropertyValue("display"),
            priority: child.style.getPropertyPriority("display"),
        }));

    let contentHeight: number;

    try {
        for (const { child } of outOfFlowDescendants) {
            child.style.setProperty("display", "none", "important");
        }

        contentHeight = element.scrollHeight + border.top + border.bottom;
    } finally {
        for (const { child, display, priority } of outOfFlowDescendants) {
            if (display) {
                child.style.setProperty("display", display, priority);
            } else {
                child.style.removeProperty("display");
            }
        }
    }

    return {
        width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,
        height: Math.max(
            sizes.reduce((total, size) => total + size.height, 0) + padding.top + padding.bottom,
            contentHeight,
        ),
    };
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
    const layout = element.closest<HTMLElement>("#splitter")?.dataset.layout;
    const definitions = TIMER_COMPONENTS.filter((definition) => {
        const component = element.querySelector<HTMLElement>(definition.selector);
        if (!component) return false;
        if (["#time-seconds", "#time-sep-sc", "#time-centis"].includes(definition.selector)) return true;
        return component.dataset.present === "1" || component.dataset.present === "true";
    });
    if (layout === "horizontal") {
        return getContainerMinimum(element, definitions, "row");
    }

    const sizes = definitions.map((definition) => getComponentMinimum(element, definition));
    const padding = getBoxPadding(element);

    return {
        width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,
        height: Math.max(0, ...sizes.map((size) => size.height)) + padding.top + padding.bottom,
    };
}

const WORLD_RECORD_PLAYER_COMPONENTS: ComponentDefinition[] = [
    {
        selector: "#world-record-label",
        widthVariable: "--splitter-world-record-label-min-width",
        heightVariable: "--splitter-world-record-label-min-height",
    },
    {
        selector: "#world-record-names",
        widthVariable: "--splitter-world-record-names-min-width",
        heightVariable: "--splitter-world-record-names-min-height",
        contentAware: {
            // width: true,
            height: true,
        },
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
    const layout = worldRecord.closest<HTMLElement>("#splitter")?.dataset.layout;
    const sizes = definitions.map((definition) => getComponentMinimum(worldRecord, definition));
    const padding = getBoxPadding(worldRecord);
    if (layout === "horizontal") {
        return {
            width: sizes.reduce((total, size) => total + size.width, 0) + padding.left + padding.right,
            height: Math.max(0, ...sizes.map((size) => size.height)) + padding.top + padding.bottom,
        };
    }
    return {
        width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,
        height: Math.max(0, ...sizes.map((size) => size.height)) + padding.top + padding.bottom,
    };
}

export function calculateWorldRecordMinimumSize(element: HTMLElement): MinimumSize {
    const sections = [
        {
            selector: "#world-record-players",
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

    const layout = element.closest<HTMLElement>("#splitter")?.dataset.layout;
    return layout === "horizontal"
        ? {
              width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,
              height: Math.max(0, ...sizes.map((size) => size.height)) + padding.top + padding.bottom,
          }
        : {
              width: Math.max(0, ...sizes.map((size) => size.width)) + padding.left + padding.right,
              height: sizes.reduce((total, size) => total + size.height, 0) + padding.top + padding.bottom,
          };
}

function combineSizesAlongAxis(sizes: MinimumSize[], direction: "row" | "column", gap: number): MinimumSize {
    return direction === "row"
        ? {
              width: sizes.reduce((total, size) => total + size.width, 0) + Math.max(0, sizes.length - 1) * gap,
              height: Math.max(0, ...sizes.map((size) => size.height)),
          }
        : {
              width: Math.max(0, ...sizes.map((size) => size.width)),
              height: sizes.reduce((total, size) => total + size.height, 0) + Math.max(0, sizes.length - 1) * gap,
          };
}

function combineGridSizes(element: HTMLElement, children: HTMLElement[], sizes: MinimumSize[]): MinimumSize {
    const columns = new Map<number, number[]>();
    const rows = new Map<number, number[]>();

    children.forEach((child, index) => {
        const rect = child.getBoundingClientRect();
        const column = Math.round(rect.left * 2) / 2;
        const row = Math.round(rect.top * 2) / 2;
        columns.set(column, [...(columns.get(column) ?? []), sizes[index].width]);
        rows.set(row, [...(rows.get(row) ?? []), sizes[index].height]);
    });

    const { row: rowGap, column: columnGap } = getGap(element);
    const width = Array.from(columns.values()).reduce((total, track) => total + Math.max(0, ...track), 0);
    const height = Array.from(rows.values()).reduce((total, track) => total + Math.max(0, ...track), 0);

    return {
        width: width + Math.max(0, columns.size - 1) * columnGap,
        height: height + Math.max(0, rows.size - 1) * rowGap,
    };
}

export function calculateSplitterInfoMinimumSize(element: HTMLElement): MinimumSize {
    const layout = element.closest<HTMLElement>("#splitter")?.dataset.layout;
    const comparison = element.querySelector<HTMLElement>(".comparison-mode");
    const timer = element.querySelector<HTMLElement>("#time-container");
    const worldRecord = element.querySelector<HTMLElement>("#world-record");

    const children: HTMLElement[] = [];
    const sizes: MinimumSize[] = [];

    if (comparison) {
        children.push(comparison);
        sizes.push(
            getEffectiveMinimumSize(comparison, "--splitter-comparison-min-width", "--splitter-comparison-min-height", {
                width: layout === "horizontal",
                height: true,
            }),
        );
    }

    if (timer) {
        children.push(timer);
        sizes.push(calculateTimeContainerMinimumSize(timer));
    }

    if (worldRecord) {
        children.push(worldRecord);
        sizes.push(calculateWorldRecordMinimumSize(worldRecord));
    }

    const padding = getBoxPadding(element);
    const style = getComputedStyle(element);
    let minimum: MinimumSize;

    if (style.display === "grid" || style.display === "inline-grid") {
        minimum = combineGridSizes(element, children, sizes);
    } else {
        const direction = style.display.includes("flex") && style.flexDirection.startsWith("row") ? "row" : "column";
        const gap = direction === "row" ? parseFloat(style.columnGap) || 0 : parseFloat(style.rowGap) || 0;
        minimum = combineSizesAlongAxis(sizes, direction, gap);
    }

    return {
        width: minimum.width + padding.left + padding.right,
        height:
            layout === "vertical"
                ? Math.max(
                      minimum.height + padding.top + padding.bottom,
                      getContentMinimumSize(element, { height: true }).height,
                  )
                : minimum.height + padding.top + padding.bottom,
    };
}
