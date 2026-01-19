import React, { JSX, useEffect, useMemo, useRef, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";
import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";
import SplitPayload from "../../models/splitPayload";
import { CompareAgainst, Comparison } from "./Splitter";
import { displayFormattedTimeParts, formatDuration, msToParts } from "./Timer";

type SplitListParameters = {
    sessionPayload: SessionPayload;
    comparison: Comparison;
};

type FlatSegment = {
    Segment: SegmentPayload;
    Depth: number;
};

type Targets = {
    cumulative: Record<string, number>;
    individual: Record<string, number>;
};

function flattenSegments(segments: SegmentPayload[], depth: number = 0): FlatSegment[] {
    const flatSegments: FlatSegment[] = [];
    for (const segment of segments) {
        flatSegments.push({ Segment: segment, Depth: depth });
        if (segment.children.length > 0) {
            flatSegments.push(...flattenSegments(segment.children, depth + 1));
        }
    }
    return flatSegments;
}

// Delta time display for splits and active row
const getDeltaDisplayTime = (delta: number, gold: boolean = false) => {
    const t = displayFormattedTimeParts(formatDuration(msToParts(delta)));
    let className = "";

    if (!gold) {
        if (delta > 0) className = "timer-behind";
        if (delta < 0) className = "timer-ahead";
    }

    return (
        <strong className={className}>
            {delta > 0 && "+"}
            {t[0]}
            <small>{t[1]}</small>
        </strong>
    );
};

// Segment time display
const getSegmentDisplayTime = (
    segment: SegmentPayload,
    split: SplitPayload | null,
    targetCumulative: number | null,
    targetIndividual: number | null,
): JSX.Element => {
    const gold = segment.gold;

    // Completed split
    if (split !== null) {
        let className = "";
        const displayTime = displayFormattedTimeParts(formatDuration(msToParts(split.current_cumulative)));

        if (gold && split.current_duration < gold) {
            className = "timer-gold";
        } else if (targetIndividual) {
            if (split.current_duration > targetIndividual) className = "timer-behind";
            if (split.current_duration < targetIndividual) className = "timer-ahead";
        }

        return <strong className={className}>{displayTime}</strong>;
    }

    // Default target display
    if (targetCumulative == null) {
        return <strong>-</strong>;
    }

    const t = displayFormattedTimeParts(formatDuration(msToParts(targetCumulative)));
    return (
        <strong>
            {t[0]}
            <small>{t[1]}</small>
        </strong>
    );
};

function segmentRow(
    segmentData: FlatSegment,
    split: SplitPayload | null,
    cumulativeTarget: number | null,
    individualTarget: number | null,
    activeRow: boolean = false,
    time: number | null = null,
    activeRowRef?: React.RefObject<HTMLTableRowElement | null>,
) {
    let delta: number | null = null;

    if (split != null && cumulativeTarget) {
        delta = split.current_cumulative - cumulativeTarget;
    } else if (activeRow && time !== null && cumulativeTarget && time > cumulativeTarget - 60000) {
        delta = time - cumulativeTarget;
    }

    return (
        <tr
            ref={activeRow ? (activeRowRef ?? null) : null}
            className={activeRow ? "selected" : ""}
            key={segmentData.Segment.id}
        >
            <td className="splitName" style={{ paddingLeft: segmentData.Depth * 16 }}>
                {segmentData.Segment.name}
            </td>

            <td className="splitDelta">{delta !== null && getDeltaDisplayTime(delta)}</td>

            <td className="splitComparison">
                {getSegmentDisplayTime(segmentData.Segment, split, cumulativeTarget, individualTarget)}
            </td>
        </tr>
    );
}

type ActiveRowProps = {
    segmentData: FlatSegment;
    cTarget: number;
    iTarget: number;
    activeRowRef: React.RefObject<HTMLTableRowElement | null>;
};

function ActiveRow({ segmentData, cTarget, iTarget, activeRowRef }: ActiveRowProps) {
    const [time, setTime] = useState(0);

    useEffect(() => {
        return EventsOn("timer:update", (val: number) => {
            setTime(val);
        });
    }, []);

    return segmentRow(segmentData, null, cTarget, iTarget, true, time, activeRowRef);
}

export default function SegmentList({ sessionPayload, comparison }: SplitListParameters) {
    const activeRowRef = useRef<HTMLTableRowElement | null>(null);
    const containerRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        const row = activeRowRef.current;
        const container = containerRef.current;

        if (!row || !container) return;

        if (!isElementFullyVisible(row, container)) {
            row.scrollIntoView({
                behavior: "smooth",
                block: "start",
            });
        }
    }, [sessionPayload.current_segment_index]);

    const targets = useMemo<Targets>(() => {
        let cumulative = 0;
        const results: Targets = { cumulative: {}, individual: {} };

        sessionPayload.leaf_segments?.forEach((segment) => {
            if (segment.average !== 0) {
                switch (comparison) {
                    case CompareAgainst.Average:
                        results.individual[segment.id] = segment.average;
                        break;
                    case CompareAgainst.Best:
                        results.individual[segment.id] = segment.pb;
                        break;
                }
                results.cumulative[segment.id] = results.individual[segment.id] + cumulative;
                cumulative += results.individual[segment.id];
            }
        });

        return results;
    }, [comparison, sessionPayload.leaf_segments]);

    const flatSegments = useMemo<FlatSegment[]>(() => {
        if (!sessionPayload.loaded_split_file) return [];
        return flattenSegments(sessionPayload.loaded_split_file.segments);
    }, [sessionPayload.loaded_split_file]);

    // Precompute leaf index lookup for O(1) membership tests
    const leafIndexById = useMemo(() => {
        const m = new Map<string, number>();
        sessionPayload.leaf_segments?.forEach((leaf, idx) => m.set(leaf.id, idx));
        return m;
    }, [sessionPayload.leaf_segments]);

    // Determine the final leaf segment id (so we can render it separately)
    const finalLeafId = useMemo(() => {
        const leaves = sessionPayload.leaf_segments;
        if (!leaves || leaves.length === 0) return null;
        return leaves[leaves.length - 1].id;
    }, [sessionPayload.leaf_segments]);

    const { mainRows, finalRow } = useMemo(() => {
        const main: JSX.Element[] = [];
        let final: JSX.Element | null = null;

        if (!sessionPayload.loaded_split_file || sessionPayload.leaf_segments == null) {
            return { mainRows: main, finalRow: final };
        }

        for (const segmentData of flatSegments) {
            const leafIndex = leafIndexById.get(segmentData.Segment.id);

            // Parent (non-leaf) segment
            if (leafIndex === undefined) {
                main.push(
                    <tr key={segmentData.Segment.id} className="parentRow">
                        <td className="splitName" style={{ paddingLeft: segmentData.Depth * 16 }}>
                            <strong>{segmentData.Segment.name}</strong>
                        </td>
                        <td className="splitDelta" />
                        <td className="splitComparison" />
                    </tr>,
                );
                continue;
            }

            const isSelected = leafIndex === sessionPayload.current_segment_index;

            const cTarget = targets.cumulative[segmentData.Segment.id];
            const iTarget = targets.individual[segmentData.Segment.id];

            const split = sessionPayload.current_run?.splits[segmentData.Segment.id] ?? null;

            const rowEl = isSelected ? (
                <ActiveRow
                    activeRowRef={activeRowRef}
                    key={segmentData.Segment.id}
                    segmentData={segmentData}
                    cTarget={cTarget}
                    iTarget={iTarget}
                />
            ) : (
                segmentRow(segmentData, split, cTarget, iTarget)
            );

            // Separate final leaf row
            if (finalLeafId && segmentData.Segment.id === finalLeafId) {
                final = rowEl;
            } else {
                main.push(rowEl);
            }
        }

        return { mainRows: main, finalRow: final };
    }, [
        sessionPayload.loaded_split_file,
        sessionPayload.leaf_segments,
        sessionPayload.current_segment_index,
        sessionPayload.current_run?.splits,
        flatSegments,
        leafIndexById,
        targets,
        finalLeafId,
    ]);

    return (
        <div className="splitList">
            <div className="gameInfo">
                <h1 className="gameTitle">
                    <strong>{sessionPayload.loaded_split_file?.game_name}</strong>
                </h1>
                <h2 className="gameCategory">
                    <small>{sessionPayload.loaded_split_file?.game_category}</small>
                </h2>
            </div>

            <div className="splitBody">
                <div ref={containerRef} className="splitContainer">
                    <table cellSpacing="0">
                        <tbody>{mainRows}</tbody>
                    </table>
                </div>

                <div className="finalSegment">
                    <table>
                        <tbody>{finalRow}</tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}

function isElementFullyVisible(element: HTMLElement, container: HTMLElement): boolean {
    const elRect = element.getBoundingClientRect();
    const containerRect = container.getBoundingClientRect();

    return elRect.top >= containerRect.top && elRect.bottom <= containerRect.bottom;
}
