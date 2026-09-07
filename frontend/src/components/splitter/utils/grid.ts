import { getBoxPadding, getGap } from "./css";
import type { MinimumSize } from "./types";

type GridPlacement = {
    row: number;
    column: number;
    rowSpan: number;
    columnSpan: number;
};

function parseGridLine(value: string, fallback: number): number {
    const match = value.match(/-?\d+/);

    if (!match) {
        return fallback;
    }

    const result = Number.parseInt(match[0], 10);

    return Number.isFinite(result) ? result : fallback;
}

function getPlacement(element: HTMLElement): GridPlacement {
    const style = getComputedStyle(element);

    const row = parseGridLine(style.gridRowStart, 1);
    const column = parseGridLine(style.gridColumnStart, 1);

    const rowEnd = parseGridLine(style.gridRowEnd, row + 1);
    const columnEnd = parseGridLine(style.gridColumnEnd, column + 1);

    return {
        row,
        column,
        rowSpan: Math.max(1, rowEnd - row),
        columnSpan: Math.max(1, columnEnd - column),
    };
}

function getTrackSizes(value: string): number[] {
    if (!value || value === "none") {
        return [];
    }

    return value.split(/\s+/).map((track) => {
        const pixels = Number.parseFloat(track);
        return Number.isFinite(pixels) ? pixels : 0;
    });
}

function distributeMinimums(
    children: Array<{
        size: MinimumSize;
        placement: GridPlacement;
    }>,
    axis: "width" | "height",
    trackCount: number,
): number[] {
    const tracks = Array.from({ length: trackCount }, () => 0);

    for (const child of children) {
        const start = axis === "width" ? child.placement.column : child.placement.row;

        const span = axis === "width" ? child.placement.columnSpan : child.placement.rowSpan;

        const size = child.size[axis];

        const index = Math.max(0, start - 1);
        const perTrack = size / span;

        for (let offset = 0; offset < span && index + offset < tracks.length; offset++) {
            tracks[index + offset] = Math.max(tracks[index + offset], perTrack);
        }
    }

    return tracks;
}

export function calculateGridMinimumSize(
    element: HTMLElement,
    children: Array<{
        element: HTMLElement;
        size: MinimumSize;
    }>,
): MinimumSize {
    const style = getComputedStyle(element);

    const columns = getTrackSizes(style.gridTemplateColumns);

    const rows = getTrackSizes(style.gridTemplateRows);

    if (columns.length === 0 || rows.length === 0) {
        return {
            width: 0,
            height: 0,
        };
    }

    const placements = children.map((child) => ({
        size: child.size,
        placement: getPlacement(child.element),
    }));

    const minimumColumns = distributeMinimums(placements, "width", columns.length);

    const minimumRows = distributeMinimums(placements, "height", rows.length);

    const gap = getGap(element);
    const padding = getBoxPadding(element);

    return {
        width:
            minimumColumns.reduce((total, size) => total + size, 0) +
            Math.max(0, columns.length - 1) * gap.column +
            padding.left +
            padding.right,
        height:
            minimumRows.reduce((total, size) => total + size, 0) +
            Math.max(0, rows.length - 1) * gap.row +
            padding.top +
            padding.bottom,
    };
}
