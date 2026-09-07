import { getBoxPadding, getEffectiveMinimumSize, getGap } from "./css";
import type { MinimumSize } from "./types";

type SegmentComponent = {
    selector: string;
    widthVariable: string;
    heightVariable: string;
    contentAware?: {
        width?: boolean;
        height?: boolean;
    };
};

/*
 * The values declared by these variables are the complete minimum
 * component sizes. The components use box-sizing: border-box, so
 * padding is already included in those dimensions.
 */
const SEGMENT_COMPONENTS: SegmentComponent[] = [
    {
        selector: ".segmentIcon",
        widthVariable: "--splitter-segment-icon-min-width",
        heightVariable: "--splitter-segment-icon-min-height",
    },
    {
        selector: ".splitName",
        widthVariable: "--splitter-segment-name-min-width",
        heightVariable: "--splitter-segment-name-min-height",
        contentAware: {
            height: true,
        },
    },
    {
        selector: ".splitDelta",
        widthVariable: "--splitter-segment-delta-min-width",
        heightVariable: "--splitter-segment-delta-min-height",
    },
    {
        selector: ".splitComparison",
        widthVariable: "--splitter-segment-comparison-min-width",
        heightVariable: "--splitter-segment-comparison-min-height",
    },
    {
        selector: ".splitTime",
        widthVariable: "--splitter-segment-time-min-width",
        heightVariable: "--splitter-segment-time-min-height",
    },
];

function getComponentMinimumSize(element: HTMLElement, definition: SegmentComponent): MinimumSize {
    return getEffectiveMinimumSize(
        element,
        definition.widthVariable,
        definition.heightVariable,
        definition.contentAware,
    );
}

/**
 * Minimum size of one segment row.
 *
 * Vertical:
 *
 *     row width  = sum(component widths)
 *     row height = max(component heights)
 *
 * Horizontal:
 *
 *     row width  = max(component widths)
 *     row height = sum(component heights)
 */
function getSegmentRowMinimumSize(row: HTMLElement, layout: string | undefined): MinimumSize {
    const components = SEGMENT_COMPONENTS.map((definition) => {
        const component = row.querySelector<HTMLElement>(definition.selector);

        return component ? getComponentMinimumSize(component, definition) : null;
    }).filter((size): size is MinimumSize => size !== null);

    if (components.length === 0) {
        return {
            width: 0,
            height: 0,
        };
    }

    if (layout === "horizontal") {
        return {
            width: Math.max(...components.map((size) => size.width)),
            height: components.reduce((total, size) => total + size.height, 0),
        };
    }

    return {
        width: components.reduce((total, size) => total + size.width, 0),
        height: Math.max(...components.map((size) => size.height)),
    };
}

/**
 * Minimum size of the tbody.
 *
 * A vertical table's rows share the same width, so the table needs
 * the largest row width. Its minimum height is the largest row
 * height because the final segment contains the one visible final
 * row.
 */
function getTbodyMinimumSize(tbody: HTMLElement, layout: string | undefined): MinimumSize {
    const rows = Array.from(tbody.querySelectorAll<HTMLElement>(":scope > tr.segmentRow, :scope > tr.parentRow"));

    if (rows.length === 0) {
        return {
            width: 0,
            height: 0,
        };
    }

    const rowSizes = rows.map((row) => getSegmentRowMinimumSize(row, layout));

    return {
        width: Math.max(...rowSizes.map((size) => size.width)),
        height: Math.max(...rowSizes.map((size) => size.height)),
    };
}

/**
 * Minimum size of the table.
 *
 * The table must be large enough to contain its tbody's required
 * component dimensions.
 */
function getTableMinimumSize(table: HTMLElement, layout: string | undefined): MinimumSize {
    const tbodies = Array.from(table.querySelectorAll<HTMLElement>(":scope > tbody"));

    if (tbodies.length === 0) {
        return {
            width: 0,
            height: 0,
        };
    }

    const tbodySizes = tbodies.map((tbody) => getTbodyMinimumSize(tbody, layout));

    return {
        width: Math.max(...tbodySizes.map((size) => size.width)),
        height: Math.max(...tbodySizes.map((size) => size.height)),
    };
}

/**
 * Minimum size of the final segment.
 *
 * Vertical:
 *
 *     finalSegment
 *       └── table
 *             └── tbody
 *                   └── segmentRow
 *
 * The final segment's table is constrained to the width of the
 * splitter. Its component minimums are used only to calculate the
 * minimum window size; CSS is still allowed to shrink the cells
 * when the table is rendered.
 */
function getFinalSegmentMinimumSize(element: HTMLElement): MinimumSize {
    const table = element.querySelector<HTMLElement>(":scope > table");

    if (!table) {
        return {
            width: 0,
            height: 0,
        };
    }

    const splitter = element.closest<HTMLElement>("#splitter");
    const layout = splitter?.dataset.layout;

    const tableMinimum = getTableMinimumSize(table, layout);
    const padding = getBoxPadding(element);

    return {
        width: tableMinimum.width + padding.left + padding.right,
        height: tableMinimum.height + padding.top + padding.bottom,
    };
}

export function calculateSplitListMinimumSize(element: HTMLElement): MinimumSize {
    const splitContainer = element.querySelector<HTMLElement>("#splitContainer");
    const finalSegment = element.querySelector<HTMLElement>("#finalSegment");

    const layout = element.closest<HTMLElement>("#splitter")?.dataset.layout;

    /*
     * Vertical:
     *
     *     splitList
     *       ├── splitContainer = flexible/scrollable
     *       └── finalSegment = intrinsic minimum
     *
     * Therefore the splitList minimum height comes from the final
     * segment.
     */
    if (layout === "vertical") {
        const finalSegmentMinimum = finalSegment ? getFinalSegmentMinimumSize(finalSegment) : { width: 0, height: 0 };

        const padding = getBoxPadding(element);

        return {
            width: finalSegmentMinimum.width + padding.left + padding.right,
            height: finalSegmentMinimum.height + padding.top + padding.bottom,
        };
    }

    /*
     * Horizontal:
     *
     *     splitContainer + finalSegment
     */
    let splitContainerMinimum: MinimumSize = {
        width: 0,
        height: 0,
    };

    const table = splitContainer?.querySelector<HTMLElement>(":scope > table");

    if (table) {
        splitContainerMinimum = getTableMinimumSize(table, layout);
    }

    const finalSegmentMinimum = finalSegment ? getFinalSegmentMinimumSize(finalSegment) : { width: 0, height: 0 };

    const gap = getGap(element);
    const padding = getBoxPadding(element);

    return {
        width: splitContainerMinimum.width + finalSegmentMinimum.width + gap.column + padding.left + padding.right,
        height: Math.max(splitContainerMinimum.height, finalSegmentMinimum.height) + padding.top + padding.bottom,
    };
}
