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
    ParentId: string | null;
    HasChildren: boolean;
};

type Targets = {
    cumulative: Record<string, number>;
    individual: Record<string, number>;
};

function flattenSegments(segments: SegmentPayload[], depth: number = 0, parentId: string | null = null): FlatSegment[] {
    const flat: FlatSegment[] = [];
    for (const segment of segments) {
        flat.push({
            Segment: segment,
            Depth: depth,
            ParentId: parentId,
            HasChildren: segment.children.length > 0,
        });

        if (segment.children.length > 0) {
            flat.push(...flattenSegments(segment.children, depth + 1, segment.id));
        }
    }
    return flat;
}

// Delta time display for splits and active row
const getDeltaDisplayTime = (delta: number, gold: boolean = false) => {
    const t = displayFormattedTimeParts(formatDuration(msToParts(delta)));
    let className = "";

    if (gold) {
        className = "timer-gold";
    } else if (delta > 0) {
        className = "timer-behind";
    } else if (delta < 0) {
        className = "timer-ahead";
    }

    return (
        <strong className={className}>
            {delta > 0 && "+"}
            {t[0]}
            <small>{t[1]}</small>
        </strong>
    );
};

const getCumulativeTimeDisplay = (split: SplitPayload | null, targetCumulative: number | null): JSX.Element => {
    const value = split?.current_cumulative ?? targetCumulative;

    if (value == null) {
        return <strong className="target">-</strong>;
    }

    const t = displayFormattedTimeParts(formatDuration(msToParts(value)));

    return (
        <strong className="target">
            {t[0]}
            <small>{t[1]}</small>
        </strong>
    );
};

const getComparisonColumn = (
    segment: SegmentPayload,
    split: SplitPayload | null,
    targetIndividual: number | null,
    runningTime: number | null,
): JSX.Element => {
    const gold = segment.gold;

    // completed segment -> show segment delta
    if (split) {
        if (targetIndividual == null) {
            return <strong>-</strong>;
        }

        const delta = split.current_duration - targetIndividual;
        return getDeltaDisplayTime(delta, gold !== 0 && split.current_duration < gold);
    }

    // active segment
    if (runningTime != null && targetIndividual != null) {
        if (runningTime >= targetIndividual - 60000) {
            return getDeltaDisplayTime(runningTime - targetIndividual);
        }
    }

    // otherwise show comparison segment time
    if (targetIndividual == null) {
        return <strong className="target">-</strong>;
    }

    const t = displayFormattedTimeParts(formatDuration(msToParts(targetIndividual)));

    return (
        <strong className="target">
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
    previousCumulative = 0,
    activeRowRef?: React.RefObject<HTMLTableRowElement | null>,
    renderToggle?: JSX.Element | null,
) {
    let delta: number | null = null;

    if (split != null && cumulativeTarget) {
        delta = split.current_cumulative - cumulativeTarget;
    } else if (activeRow && time !== null && cumulativeTarget && time > cumulativeTarget - 60000) {
        delta = time - cumulativeTarget;
    }

    const runningSegmentTime = activeRow && time != null ? time - previousCumulative : null;

    return (
        <tr
            ref={activeRow ? (activeRowRef ?? null) : null}
            className={"segmentRow" + (activeRow ? " selected" : "")}
            key={segmentData.Segment.id}
        >
            <td className="splitName" style={{ paddingLeft: 5 + segmentData.Depth * 16 }}>
                {renderToggle}

                {segmentData.Segment.icon && (
                    <img
                        src={segmentData.Segment.icon}
                        alt=""
                        draggable={false}
                        style={{
                            width: 18,
                            height: 18,
                            objectFit: "contain",
                            marginRight: 6,
                            verticalAlign: "middle",
                            borderRadius: 2,
                        }}
                    />
                )}

                {segmentData.Segment.name}
            </td>

            <td className="splitDelta">{delta !== null && getDeltaDisplayTime(delta)}</td>

            <td className="splitComparison">
                {getComparisonColumn(segmentData.Segment, split, individualTarget, runningSegmentTime)}
            </td>

            <td className="splitTime">{getCumulativeTimeDisplay(split, cumulativeTarget)}</td>
        </tr>
    );
}

type ActiveRowProps = {
    segmentData: FlatSegment;
    cTarget: number;
    iTarget: number;
    previousCumulative: number;
    activeRowRef: React.RefObject<HTMLTableRowElement | null>;
};

function ActiveRow({ segmentData, cTarget, iTarget, previousCumulative, activeRowRef }: ActiveRowProps) {
    const [time, setTime] = useState(0);

    useEffect(() => {
        return EventsOn("timer:update", (val: number) => {
            setTime(val);
        });
    }, []);

    return segmentRow(segmentData, null, cTarget, iTarget, true, time, previousCumulative, activeRowRef);
}

function isElementFullyVisible(element: HTMLElement, container: HTMLElement): boolean {
    const elRect = element.getBoundingClientRect();
    const containerRect = container.getBoundingClientRect();

    return elRect.top >= containerRect.top && elRect.bottom <= containerRect.bottom;
}

function getAncestorIds(leafId: string, parentById: Map<string, string | null>): string[] {
    const ancestors: string[] = [];
    let cur: string | null | undefined = leafId;

    while (cur != null) {
        const p = parentById.get(cur);
        if (p == null) break;
        ancestors.push(p);
        cur = p;
    }

    return ancestors;
}

function isVisible(id: string, parentById: Map<string, string | null>, expandedParents: Set<string>): boolean {
    let cur: string | null | undefined = id;
    while (cur != null) {
        const p = parentById.get(cur);
        if (p == null) return true; // reached root
        if (!expandedParents.has(p)) return false;
        cur = p;
    }
    return true;
}

export default function SegmentList({ sessionPayload, comparison }: SplitListParameters) {
    const [completeClassName, setCompleteClassName] = React.useState<string>("");
    const activeRowRef = useRef<HTMLTableRowElement | null>(null);
    const containerRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        let className = "";
        if (sessionPayload.loaded_split_file && sessionPayload.current_run && sessionPayload.leaf_segments) {
            if (Object.keys(sessionPayload.current_run.splits).length == sessionPayload.leaf_segments.length) {
                className = "complete";

                const pb = sessionPayload.loaded_split_file.pb;
                console.log(sessionPayload.loaded_split_file);
                if (pb) {
                    const segments = sessionPayload.leaf_segments;
                    if (segments) {
                        const finalSplit = segments[segments.length - 1].id;
                        const finalTime = sessionPayload.current_run.splits[finalSplit].current_cumulative;
                        if (finalTime < pb.total_time) {
                            className += " pb";
                        }
                    }
                }
            }
            setCompleteClassName(className);
        } else {
            setCompleteClassName("");
        }
    }, [sessionPayload]);

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
                    case CompareAgainst.SumOfBest:
                        results.individual[segment.id] = segment.gold;
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

    const parentById = useMemo(() => {
        const m = new Map<string, string | null>();
        for (const fs of flatSegments) {
            m.set(fs.Segment.id, fs.ParentId);
        }
        return m;
    }, [flatSegments]);
    const [expandedParents, setExpandedParents] = useState<Set<string>>(() => new Set());

    // For each parent segment id, find the "last" leaf (by leaf_segments order) in its subtree.
    const lastLeafByParentId = useMemo(() => {
        const result = new Map<string, string>(); // parentId -> leafId

        if (!sessionPayload.leaf_segments || sessionPayload.leaf_segments.length === 0) return result;

        // For each leaf, walk ancestors and assign/overwrite (later leaves overwrite earlier => "last" wins)
        for (const leaf of sessionPayload.leaf_segments) {
            const ancestors = getAncestorIds(leaf.id, parentById);
            for (const anc of ancestors) {
                result.set(anc, leaf.id);
            }
        }

        return result;
    }, [sessionPayload.leaf_segments, parentById]);

    const leavesByParentId = useMemo(() => {
        const map = new Map<string, SegmentPayload[]>();

        for (const leaf of sessionPayload.leaf_segments ?? []) {
            for (const parent of getAncestorIds(leaf.id, parentById)) {
                const existing = map.get(parent);
                if (existing) {
                    existing.push(leaf);
                } else {
                    map.set(parent, [leaf]);
                }
            }
        }

        return map;
    }, [sessionPayload.leaf_segments, parentById]);

    // Determine the final leaf segment id (so we can render it separately)
    const finalLeafId = useMemo(() => {
        const leaves = sessionPayload.leaf_segments;
        if (!leaves || leaves.length === 0) return null;
        return leaves[leaves.length - 1].id;
    }, [sessionPayload.leaf_segments]);

    useEffect(() => {
        const leaves = sessionPayload.leaf_segments;
        if (!leaves || leaves.length === 0) {
            setExpandedParents(new Set());
            return;
        }

        const activeLeaf = leaves[sessionPayload.current_segment_index];
        if (!activeLeaf) {
            setExpandedParents(new Set());
            return;
        }

        const ancestors = getAncestorIds(activeLeaf.id, parentById);
        setExpandedParents(new Set(ancestors));
    }, [sessionPayload.current_segment_index, sessionPayload.leaf_segments, parentById]);

    // Keep active row in view
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

    const toggleParent = (parentId: string) => {
        setExpandedParents((prev) => {
            const next = new Set(prev);
            if (next.has(parentId)) next.delete(parentId);
            else next.add(parentId);
            return next;
        });
    };

    const { mainRows, finalRow } = useMemo(() => {
        const main: JSX.Element[] = [];
        let final: JSX.Element | null = null;

        if (!sessionPayload.loaded_split_file || sessionPayload.leaf_segments == null) {
            return { mainRows: main, finalRow: final };
        }

        for (const segmentData of flatSegments) {
            const isFinalLeaf = finalLeafId != null && segmentData.Segment.id === finalLeafId;

            // Hide collapsed rows, BUT always allow the final leaf through
            if (!isFinalLeaf && !isVisible(segmentData.Segment.id, parentById, expandedParents)) {
                continue;
            }

            const leafIndex = leafIndexById.get(segmentData.Segment.id);

            // Parent (non-leaf) segment
            if (leafIndex === undefined) {
                const isExpanded = expandedParents.has(segmentData.Segment.id);
                const hasChildren = segmentData.HasChildren;

                const toggle = hasChildren ? (
                    <button
                        type="button"
                        className="collapseToggle"
                        onClick={(e) => {
                            e.stopPropagation();
                            toggleParent(segmentData.Segment.id);
                        }}
                        aria-label={isExpanded ? "Collapse segment group" : "Expand segment group"}
                    >
                        {isExpanded ? "▾" : "▸"}
                    </button>
                ) : null;

                const lastLeafId = lastLeafByParentId.get(segmentData.Segment.id) ?? null;
                const lastLeafSplit = lastLeafId ? (sessionPayload.current_run?.splits[lastLeafId] ?? null) : null;

                // Comparison time (cumulative display) pulled from last leaf
                let parentComparison: JSX.Element | null = null;

                // Delta pulled from last leaf vs its cumulative target
                let parentDelta: JSX.Element | null = null;

                let parentSegmentDelta: JSX.Element | null = null;

                if (lastLeafId) {
                    const cTarget = targets.cumulative[lastLeafId] ?? null;

                    // Parent cumulative time
                    parentComparison = getCumulativeTimeDisplay(lastLeafSplit, cTarget);

                    // Parent cumulative delta
                    if (lastLeafSplit && cTarget != null) {
                        parentDelta = getDeltaDisplayTime(lastLeafSplit.current_cumulative - cTarget);
                    }

                    // Parent total segment delta
                    let actual = 0;
                    let target = 0;
                    let complete = true;

                    for (const leaf of leavesByParentId.get(segmentData.Segment.id) ?? []) {
                        const split = sessionPayload.current_run?.splits[leaf.id];

                        if (!split) {
                            complete = false;
                            break;
                        }

                        actual += split.current_duration;
                        target += targets.individual[leaf.id] ?? 0;
                    }

                    if (complete) {
                        parentSegmentDelta = getDeltaDisplayTime(actual - target);
                    }
                }

                main.push(
                    <tr key={segmentData.Segment.id} className="parentRow">
                        <td
                            className={"splitName " + completeClassName}
                            style={{ paddingLeft: segmentData.Depth * 16 }}
                        >
                            {toggle}

                            {segmentData.Segment.icon && (
                                <img
                                    src={segmentData.Segment.icon}
                                    alt=""
                                    draggable={false}
                                    style={{
                                        width: 18,
                                        height: 18,
                                        objectFit: "contain",
                                        marginRight: 6,
                                        verticalAlign: "middle",
                                        borderRadius: 2,
                                    }}
                                />
                            )}

                            <strong>{segmentData.Segment.name}</strong>
                        </td>

                        <td className={"splitDelta " + completeClassName}>{parentDelta}</td>

                        <td className={"splitComparison " + completeClassName}>{parentSegmentDelta}</td>

                        <td className={"splitTime " + completeClassName}>{parentComparison}</td>
                    </tr>,
                );
                continue;
            }

            const isSelected = leafIndex === sessionPayload.current_segment_index;

            const cTarget = targets.cumulative[segmentData.Segment.id];
            const iTarget = targets.individual[segmentData.Segment.id];

            const split = sessionPayload.current_run?.splits[segmentData.Segment.id] ?? null;

            const previousActual =
                leafIndex === 0
                    ? 0
                    : (sessionPayload.current_run?.splits[sessionPayload.leaf_segments![leafIndex - 1].id]
                          ?.current_cumulative ?? 0);

            const rowEl = isSelected ? (
                <ActiveRow
                    key={segmentData.Segment.id}
                    activeRowRef={activeRowRef}
                    segmentData={segmentData}
                    cTarget={cTarget}
                    iTarget={iTarget}
                    previousCumulative={previousActual}
                />
            ) : (
                segmentRow(segmentData, split, cTarget, iTarget)
            );

            // Separate final leaf row (still respects collapse via isVisible above)
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
        parentById,
        expandedParents,
    ]);

    return (
        <div id="splitList" className={completeClassName}>
            <div id="gameInfo" className={completeClassName}>
                <h1 id="gameTitle" className={completeClassName}>
                    <strong>{sessionPayload.loaded_split_file?.game_name}</strong>
                </h1>
                <h2 id="gameCategory" className={completeClassName}>
                    <small>{sessionPayload.loaded_split_file?.game_category}</small>
                </h2>
                <div id="attempts" className={completeClassName}>
                    {sessionPayload.loaded_split_file?.attempts}
                </div>
            </div>

            <div id="splitBody" className={completeClassName}>
                <div ref={containerRef} id="splitContainer" className={completeClassName}>
                    <table cellSpacing="0" className={completeClassName}>
                        <tbody>{mainRows}</tbody>
                    </table>
                </div>

                <div id="finalSegment" className={completeClassName}>
                    <table className={completeClassName}>
                        <tbody>{finalRow}</tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
