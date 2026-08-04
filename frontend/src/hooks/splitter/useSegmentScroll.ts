import { useEffect, useRef } from "react";

import { isElementFullyVisible } from "../../components/splitter/segments/segmentUtils";

export function useSegmentScroll(activeIndex: number) {
    const containerRef = useRef<HTMLDivElement>(null);

    const activeRowRef = useRef<HTMLTableRowElement>(null);

    /*
     * Scroll active row into view
     */
    useEffect(() => {
        const row = activeRowRef.current;
        const container = containerRef.current;

        if (!row || !container) {
            return;
        }

        if (!isElementFullyVisible(row, container)) {
            row.scrollIntoView({
                behavior: "smooth",
                block: "start",
            });
        }
    }, [activeIndex]);

    return {
        containerRef,
        activeRowRef,
    };
}
