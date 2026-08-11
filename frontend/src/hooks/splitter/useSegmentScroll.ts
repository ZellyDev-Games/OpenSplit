import { useEffect, useRef } from "react";

import { isElementFullyVisible } from "../../components/splitter/segments/segmentUtils";

export function useSegmentScroll(activeIndex: number) {
    const containerRef = useRef<HTMLDivElement>(null);

    const activeRowRef = useRef<HTMLTableRowElement>(null);

    /*
     * Scroll the active segment into view.
     *
     * The split list has two different scrolling directions:
     *
     *  - vertical layout   -> scrollTop
     *  - horizontal layout -> scrollLeft
     *
     * We explicitly calculate the required scroll position because
     * horizontal mode turns each table row into a segment column.
     */
    useEffect(() => {
        const row = activeRowRef.current;
        const container = containerRef.current;

        if (!row || !container) {
            return;
        }

        if (isElementFullyVisible(row, container)) {
            return;
        }

        const splitter = container.closest("#splitter");

        if (!splitter) {
            return;
        }

        const layout = splitter.getAttribute("data-layout");

        const containerRect = container.getBoundingClientRect();
        const rowRect = row.getBoundingClientRect();

        if (layout === "horizontal") {
            const rowLeft = rowRect.left - containerRect.left + container.scrollLeft;
            const rowRight = rowLeft + rowRect.width;

            const visibleLeft = container.scrollLeft;
            const visibleRight = container.scrollLeft + container.clientWidth;

            if (rowLeft < visibleLeft) {
                container.scrollTo({
                    left: Math.max(0, rowLeft - 12),
                    behavior: "smooth",
                });
            } else if (rowRight > visibleRight) {
                container.scrollTo({
                    left: Math.max(0, rowRight - container.clientWidth + 12),
                    behavior: "smooth",
                });
            }

            return;
        }

        /*
         * Vertical layout.
         */
        const rowTop = rowRect.top - containerRect.top + container.scrollTop;
        const rowBottom = rowTop + rowRect.height;

        const visibleTop = container.scrollTop;
        const visibleBottom = container.scrollTop + container.clientHeight;

        if (rowTop < visibleTop) {
            container.scrollTo({
                top: Math.max(0, rowTop - 12),
                behavior: "smooth",
            });
        } else if (rowBottom > visibleBottom) {
            container.scrollTo({
                top: Math.max(0, rowBottom - container.clientHeight + 12),
                behavior: "smooth",
            });
        }
    }, [activeIndex]);

    return {
        containerRef,
        activeRowRef,
    };
}
