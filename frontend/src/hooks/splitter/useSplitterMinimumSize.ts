import { useEffect } from "react";

import { WindowSetMinSize } from "../../../wailsjs/runtime/runtime";
import { getMinimum } from "../../components/splitter/utils/css";
import { calculateSplitterMinimumSize } from "../../components/splitter/utils/splitter";
import { log } from "../../utils/logger";

type MinimumSize = {
    width: number;
    height: number;
};

function getLayoutInfo(element: HTMLElement) {
    const style = getComputedStyle(element);

    return {
        display: style.display,
        flexDirection: style.flexDirection,
        flexWrap: style.flexWrap,
        gap: {
            column: style.columnGap,
            row: style.rowGap,
        },
        gridTemplateColumns: style.gridTemplateColumns,
        gridTemplateRows: style.gridTemplateRows,
        height: element.getBoundingClientRect().height,
        minHeight: style.minHeight,
        minWidth: style.minWidth,
        padding: {
            left: style.paddingLeft,
            right: style.paddingRight,
            top: style.paddingTop,
            bottom: style.paddingBottom,
        },
        scrollHeight: element.scrollHeight,
        scrollWidth: element.scrollWidth,
        width: element.getBoundingClientRect().width,
    };
}

function getSegmentMinimumInfo(element: HTMLElement) {
    const finalRow = element.querySelector<HTMLElement>("#finalSegment tr.segmentRow");

    if (!finalRow) {
        return null;
    }

    const definitions = [
        {
            name: "icon",
            selector: ".segmentIcon",
            width: "--splitter-segment-icon-min-width",
            height: "--splitter-segment-icon-min-height",
        },
        {
            name: "name",
            selector: ".splitName",
            width: "--splitter-segment-name-min-width",
            height: "--splitter-segment-name-min-height",
        },
        {
            name: "delta",
            selector: ".splitDelta",
            width: "--splitter-segment-delta-min-width",
            height: "--splitter-segment-delta-min-height",
        },
        {
            name: "comparison",
            selector: ".splitComparison",
            width: "--splitter-segment-comparison-min-width",
            height: "--splitter-segment-comparison-min-height",
        },
        {
            name: "time",
            selector: ".splitTime",
            width: "--splitter-segment-time-min-width",
            height: "--splitter-segment-time-min-height",
        },
    ];

    return definitions.map((definition) => {
        const component = finalRow.querySelector<HTMLElement>(definition.selector);

        if (!component) {
            return {
                name: definition.name,
                width: 0,
                height: 0,
                found: false,
            };
        }

        return {
            name: definition.name,
            width: getMinimum(component, definition.width),
            height: getMinimum(component, definition.height),
            found: true,
            cssMinWidth: getComputedStyle(component).minWidth,
            cssMinHeight: getComputedStyle(component).minHeight,
        };
    });
}

export function useSplitterMinimumSize(splitterRef: React.RefObject<HTMLDivElement | null>) {
    useEffect(() => {
        const element = splitterRef.current;

        if (!element) {
            log.debug("[SplitterMinimumSize] No splitter element");
            return;
        }

        let frame = 0;
        let updateCount = 0;
        let previousMinimum: MinimumSize | null = null;
        let lastUpdateReason = "initial";

        const update = (reason = "unknown") => {
            lastUpdateReason = reason;

            cancelAnimationFrame(frame);

            frame = requestAnimationFrame(() => {
                updateCount++;

                const minimum = calculateSplitterMinimumSize(element);

                const actual = {
                    width: element.getBoundingClientRect().width,
                    height: element.getBoundingClientRect().height,
                    scrollWidth: element.scrollWidth,
                    scrollHeight: element.scrollHeight,
                };

                const changed =
                    previousMinimum === null ||
                    previousMinimum.width !== minimum.width ||
                    previousMinimum.height !== minimum.height;

                log.debug(`[SplitterMinimumSize] ${reason} — minimum ${changed ? "changed" : "unchanged"}`);

                log.debug("[SplitterMinimumSize] Calculation", {
                    updateCount,
                    reason,
                    minimum,
                    previousMinimum,
                    changed,
                });

                log.debug("[SplitterMinimumSize] Actual splitter", actual);

                log.debug("[SplitterMinimumSize] Layout", {
                    splitter: getLayoutInfo(element),
                });

                log.debug("[SplitterMinimumSize] Component minimums", {
                    segment: getSegmentMinimumInfo(element),
                });

                if (!changed) {
                    return;
                }

                previousMinimum = minimum;

                log.debug("[SplitterMinimumSize] WindowSetMinSize", {
                    width: minimum.width,
                    height: minimum.height,
                });

                log.debug("[SplitterMinimumSize] FINAL", {
                    windowMinimum: minimum,
                    splitListStyleMinHeight: element.querySelector<HTMLElement>("#splitList")?.style.minHeight,
                    splitListRect: element.querySelector<HTMLElement>("#splitList")?.getBoundingClientRect().height,
                    finalSegmentRect: element.querySelector<HTMLElement>("#finalSegment")?.getBoundingClientRect()
                        .height,
                });

                WindowSetMinSize(minimum.width, minimum.height);
            });
        };

        const resizeObserver = new ResizeObserver((entries) => {
            log.debug(
                "[SplitterMinimumSize] Resize",
                entries.map((entry) => {
                    const target = entry.target as HTMLElement;

                    return {
                        id: target.id,
                        className: target.className,
                        width: target.getBoundingClientRect().width,
                        height: target.getBoundingClientRect().height,
                    };
                }),
            );

            update("resize");
        });

        const observedElements = new Set<HTMLElement>();

        const observe = (target: Element) => {
            if (!(target instanceof HTMLElement)) {
                return;
            }

            if (observedElements.has(target)) {
                return;
            }

            observedElements.add(target);
            resizeObserver.observe(target);
        };

        const observeResizeTargets = () => {
            observe(element);

            element
                .querySelectorAll<HTMLElement>(
                    [
                        "#gameInfo",
                        "#splitList",
                        "#splitContainer",
                        "#finalSegment",
                        "#finalSegment table",
                        "#finalSegment tbody",
                        "#finalSegment tr",
                        "#splitterInfo",
                        ".splitName",
                    ].join(", "),
                )
                .forEach(observe);
        };

        observeResizeTargets();

        const mutationObserver = new MutationObserver((mutations) => {
            const reasons = mutations.map((mutation) => {
                if (mutation.type === "childList") {
                    return "childList";
                }

                if (mutation.type === "attributes") {
                    return `${mutation.attributeName ?? "attribute"} changed`;
                }

                return mutation.type;
            });

            observeResizeTargets();

            log.debug("[SplitterMinimumSize] Mutation", {
                reasons,
                mutations,
            });

            update("mutation");
        });

        mutationObserver.observe(element, {
            childList: true,
            subtree: true,
            attributes: true,
            attributeFilter: ["class", "style", "data-layout"],
        });

        log.debug("[SplitterMinimumSize] Initialized", {
            layout: element.dataset.layout ?? "unknown",
            childCount: element.children.length,
            observedCount: observedElements.size,
        });

        update("initial");

        return () => {
            cancelAnimationFrame(frame);

            mutationObserver.disconnect();
            resizeObserver.disconnect();

            log.debug("[SplitterMinimumSize] Cleanup", {
                layout: element.dataset.layout ?? "unknown",
                updateCount,
                lastUpdateReason,
                previousMinimum,
            });
        };
    }, [splitterRef]);
}
