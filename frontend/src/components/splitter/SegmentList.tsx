/**
 * Displays the active split list during a run.
 *
 * Features:
 *  - Hierarchical segment display
 *  - Expand/collapse parent segments
 *  - Live timer updates
 *  - Comparison against Average / PB / Sum of Best
 *  - Automatic scrolling to the active split
 *  - Parent aggregate timing and delta calculations
 */

import { useEffect, useMemo, useRef, useState } from "react";

import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";
import { log } from "../../utils/logger";
import { FlatSegment, flattenSegments, getAncestorIds, isElementFullyVisible, Targets } from "./segmentUtils";
import SplitGameInfo from "./SplitGameInfo";
import { CompareAgainst, Comparison } from "./Splitter";
import { useSegmentRows } from "./useSegmentRows";

type SegmentListParameters = {
    sessionPayload: SessionPayload;
    comparison: Comparison;
    forceExpandAll?: boolean;
};

export default function SegmentList({ sessionPayload, comparison, forceExpandAll }: SegmentListParameters) {
    const [completeClassName, setCompleteClassName] = useState("");
    const activeRowRef = useRef<HTMLTableRowElement | null>(null);
    const containerRef = useRef<HTMLDivElement | null>(null);

    /*
     * Completion state
     */
    useEffect(() => {
        let className = "";

        const splitFile = sessionPayload.loaded_split_file;
        const run = sessionPayload.current_run;
        const leaves = sessionPayload.leaf_segments;

        if (splitFile && run && leaves && Object.keys(run.splits).length === leaves.length) {
            className = "complete";

            const pb = splitFile.pb;

            log.info("[SegmentList] Run completed", {
                pb: className.includes("pb"),
                totalSplits: leaves.length,
            });
            if (pb) {
                const finalSplit = leaves[leaves.length - 1];
                const finalTime = run.splits[finalSplit.id].current_cumulative;

                if (finalTime < pb.total_time) {
                    className += " pb";
                }
            }

            log.info("[SegmentList] Run completed", {
                totalSplits: leaves.length,
                pb: className.includes("pb"),
            });
        }

        log.debug(sessionPayload.loaded_split_file);
        log.debug(sessionPayload.loaded_split_file?.variables);
        setCompleteClassName(className);
    }, [sessionPayload]);

    /*
     * Comparison targets
     */
    const targets = useMemo<Targets>(() => {
        let cumulative = 0;

        const result: Targets = {
            cumulative: {},
            individual: {},
        };

        sessionPayload.leaf_segments?.forEach((segment) => {
            let value = 0;

            switch (comparison) {
                case CompareAgainst.Average:
                    value = segment.average;
                    break;

                case CompareAgainst.Best:
                    value = segment.pb;
                    break;

                case CompareAgainst.SumOfBest:
                    value = segment.gold;
                    break;
            }

            result.individual[segment.id] = value;

            result.cumulative[segment.id] = cumulative + value;

            cumulative += value;
        });

        return result;
    }, [comparison, sessionPayload.leaf_segments]);

    /*
     * Segment tree
     */
    const flatSegments = useMemo<FlatSegment[]>(() => {
        if (!sessionPayload.loaded_split_file) {
            return [];
        }

        return flattenSegments(sessionPayload.loaded_split_file.segments);
    }, [sessionPayload.loaded_split_file]);

    // Precompute leaf index lookup for O(1) membership tests
    const leafIndexById = useMemo(() => {
        const map = new Map<string, number>();

        sessionPayload.leaf_segments?.forEach((leaf, index) => {
            map.set(leaf.id, index);
        });

        return map;
    }, [sessionPayload.leaf_segments]);

    const parentById = useMemo(() => {
        const map = new Map<string, string | null>();

        flatSegments.forEach((segment) => {
            map.set(segment.Segment.id, segment.ParentId);
        });

        return map;
    }, [flatSegments]);

    const [expandedParents, setExpandedParents] = useState<Set<string>>(() => {
        const expanded = new Set<string>();

        flatSegments.forEach((segment) => {
            if (segment.Segment.children.length > 0) {
                expanded.add(segment.Segment.id);
            }
        });

        return expanded;
    });

    useEffect(() => {
        setExpandedParents(
            new Set(
                flatSegments
                    .filter((segment) => segment.Segment.children.length > 0)
                    .map((segment) => segment.Segment.id),
            ),
        );
    }, [flatSegments]);

    // For each parent segment id, find the "last" leaf (by leaf_segments order) in its subtree.
    const lastLeafByParentId = useMemo(() => {
        const result = new Map<string, string>();

        // For each leaf, walk ancestors and assign/overwrite (later leaves overwrite earlier => "last" wins)
        for (const leaf of sessionPayload.leaf_segments ?? []) {
            for (const parent of getAncestorIds(leaf.id, parentById)) {
                result.set(parent, leaf.id);
            }
        }

        return result;
    }, [sessionPayload.leaf_segments, parentById]);

    const leavesByParentId = useMemo(() => {
        const result = new Map<string, SegmentPayload[]>();

        for (const leaf of sessionPayload.leaf_segments ?? []) {
            for (const parent of getAncestorIds(leaf.id, parentById)) {
                const existing = result.get(parent);

                if (existing) {
                    existing.push(leaf);
                } else {
                    result.set(parent, [leaf]);
                }
            }
        }

        return result;
    }, [sessionPayload.leaf_segments, parentById]);

    // Determine the final leaf segment id (so we can render it separately)
    const finalLeafId = useMemo(() => {
        const leaves = sessionPayload.leaf_segments;

        if (!leaves?.length) {
            return null;
        }

        return leaves[leaves.length - 1].id;
    }, [sessionPayload.leaf_segments]);

    /*
     * Automatically expand active segment parents
     */
    useEffect(() => {
        if (forceExpandAll) {
            setExpandedParents(
                new Set(
                    flatSegments
                        .filter((segment) => segment.Segment.children.length > 0)
                        .map((segment) => segment.Segment.id),
                ),
            );
            return;
        }

        const leaves = sessionPayload.leaf_segments;

        if (!leaves) {
            setExpandedParents(new Set());
            return;
        }

        const active = leaves[sessionPayload.current_segment_index];

        if (!active) {
            return;
        }

        setExpandedParents(new Set(getAncestorIds(active.id, parentById)));
    }, [forceExpandAll, flatSegments, sessionPayload.current_segment_index, sessionPayload.leaf_segments, parentById]);

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
    }, [sessionPayload.current_segment_index]);

    const toggleParent = (id: string) => {
        setExpandedParents((previous) => {
            log.debug("[SegmentList] Toggle parent", {
                id,
                expanded: !previous.has(id),
            });
            const next = new Set(previous);

            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }

            return next;
        });
    };

    const { rows, finalRow } = useSegmentRows({
        sessionPayload,
        flatSegments,
        targets,
        completeClassName,
        leafIndexById,
        parentById,
        expandedParents,
        lastLeafByParentId,
        leavesByParentId,
        finalLeafId,
        activeRowRef,
        toggleParent,
    });

    return (
        <div id="splitList" className={completeClassName}>
            <SplitGameInfo sessionPayload={sessionPayload} completeClassName={completeClassName} />

            <div id="splitBody" className={completeClassName}>
                <div ref={containerRef} id="splitContainer" className={completeClassName}>
                    <table cellSpacing={0} className={completeClassName}>
                        <tbody>{rows}</tbody>
                    </table>
                </div>

                <div id="finalSegment" className={completeClassName}>
                    <table cellSpacing={0} className={completeClassName}>
                        <tbody>{finalRow}</tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
