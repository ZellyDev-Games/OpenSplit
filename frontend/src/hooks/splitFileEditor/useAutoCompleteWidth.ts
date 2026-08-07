import { useMemo } from "react";

export function useAutoCompleteWidth<T>(items: T[], getName: (item: T) => string, fallback = 150) {
    return useMemo(() => {
        const canvas = document.createElement("canvas");
        const ctx = canvas.getContext("2d");

        if (!ctx) {
            return fallback;
        }

        ctx.font = getComputedStyle(document.body).font;

        return Math.max(fallback, ...items.map((item) => ctx.measureText(getName(item)).width + 10));
    }, [items, fallback]);
}
