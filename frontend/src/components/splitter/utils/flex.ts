import { getBoxPadding, getGap } from "./css";
import type { MinimumSize } from "./types";

export function calculateFlexMinimumSize(element: HTMLElement, children: MinimumSize[]): MinimumSize {
    const style = getComputedStyle(element);

    const direction = style.flexDirection;
    const wrap = style.flexWrap;

    const gap = getGap(element);
    const padding = getBoxPadding(element);

    const isRow = direction === "row" || direction === "row-reverse";
    const isWrapped = wrap === "wrap" || wrap === "wrap-reverse";

    if (isRow) {
        if (isWrapped) {
            return {
                width: Math.max(0, ...children.map((child) => child.width)) + padding.left + padding.right,
                height:
                    children.reduce((total, child) => total + child.height, 0) +
                    Math.max(0, children.length - 1) * gap.row +
                    padding.top +
                    padding.bottom,
            };
        }

        return {
            width:
                children.reduce((total, child) => total + child.width, 0) +
                Math.max(0, children.length - 1) * gap.column +
                padding.left +
                padding.right,

            height: Math.max(0, ...children.map((child) => child.height)) + padding.top + padding.bottom,
        };
    }

    if (isWrapped) {
        return {
            width:
                children.reduce((total, child) => total + child.width, 0) +
                Math.max(0, children.length - 1) * gap.column +
                padding.left +
                padding.right,

            height: Math.max(0, ...children.map((child) => child.height)) + padding.top + padding.bottom,
        };
    }

    return {
        width: Math.max(0, ...children.map((child) => child.width)) + padding.left + padding.right,

        height:
            children.reduce((total, child) => total + child.height, 0) +
            Math.max(0, children.length - 1) * gap.row +
            padding.top +
            padding.bottom,
    };
}
