import { ReactNode, RefObject, useMemo } from "react";

import ActiveSegmentRow from "../../components/splitter/segments/ActiveSegmentRow";
import ParentSegmentRow from "../../components/splitter/segments/ParentSegmentRow";
import SegmentRow from "../../components/splitter/segments/SegmentRow";
import { CumulativeTimeDisplay, DeltaDisplay } from "../../components/splitter/segments/SegmentTime";
import { FlatSegment, isVisible, Targets } from "../../components/splitter/segments/segmentUtils";
import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";

type SegmentRowsResult = {
    rows: ReactNode[];
    finalRow: ReactNode | null;
};

type UseSegmentRowsParams = {
    sessionPayload: SessionPayload;
    flatSegments: FlatSegment[];
    targets: Targets;
    completeClassName: string;
    leafIndexById: Map<string, number>;
    parentById: Map<string, string | null>;
    expandedParents: Set<string>;
    lastLeafByParentId: Map<string, string>;
    leavesByParentId: Map<string, SegmentPayload[]>;
    finalLeafId: string | null;
    activeRowRef: RefObject<HTMLTableRowElement | null>;
    toggleParent: (id: string) => void;
};
export function useSegmentRows({
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
}: UseSegmentRowsParams): SegmentRowsResult {
    return useMemo<SegmentRowsResult>(() => {
        const rows: ReactNode[] = [];
        let finalRow: ReactNode = null;

        if (!sessionPayload.loaded_split_file || !sessionPayload.leaf_segments) {
            return { rows, finalRow };
        }

        for (const segmentData of flatSegments) {
            const segment = segmentData.segment;

            const isFinalLeaf = segment.id === finalLeafId;

            if (!isFinalLeaf && !isVisible(segment.id, parentById, expandedParents)) {
                continue;
            }

            const leafIndex = leafIndexById.get(segment.id);

            if (leafIndex === undefined) {
                const isExpanded = expandedParents.has(segment.id);

                const lastLeafId = lastLeafByParentId.get(segment.id) ?? null;

                const lastLeafSplit = lastLeafId ? (sessionPayload.current_run?.splits[lastLeafId] ?? null) : null;

                let parentComparison: ReactNode = null;
                let parentDelta: ReactNode = null;
                let parentSegmentDelta: ReactNode = null;

                if (lastLeafId) {
                    const cTarget = targets.cumulative[lastLeafId] ?? null;

                    parentComparison = <CumulativeTimeDisplay split={lastLeafSplit} targetCumulative={cTarget} />;

                    if (lastLeafSplit && cTarget != null) {
                        parentDelta = <DeltaDisplay delta={lastLeafSplit.current_cumulative - cTarget} />;
                    }

                    let actual = 0;
                    let target = 0;
                    let complete = true;

                    for (const leaf of leavesByParentId.get(segment.id) ?? []) {
                        const split = sessionPayload.current_run?.splits[leaf.id];

                        if (!split) {
                            complete = false;
                            break;
                        }

                        actual += split.current_duration;
                        target += targets.individual[leaf.id] ?? 0;
                    }

                    if (complete) {
                        parentSegmentDelta = <DeltaDisplay delta={actual - target} />;
                    }
                }

                rows.push(
                    <ParentSegmentRow
                        key={segment.id}
                        segment={segment}
                        depth={segmentData.depth}
                        completeClassName={completeClassName}
                        isExpanded={isExpanded}
                        hasChildren={segmentData.hasChildren}
                        parentComparison={parentComparison}
                        parentDelta={parentDelta}
                        parentSegmentDelta={parentSegmentDelta}
                        onToggle={() => toggleParent(segment.id)}
                    />,
                );

                continue;
            }

            const selected = leafIndex === sessionPayload.current_segment_index;

            const split = sessionPayload.current_run?.splits[segment.id] ?? null;

            const previousActual =
                leafIndex === 0
                    ? 0
                    : (sessionPayload.current_run?.splits[sessionPayload.leaf_segments[leafIndex - 1].id]
                          ?.current_cumulative ?? 0);

            const row = selected ? (
                <ActiveSegmentRow
                    key={segment.id}
                    activeRowRef={activeRowRef}
                    segmentData={segmentData}
                    cTarget={targets.cumulative[segment.id]}
                    iTarget={targets.individual[segment.id]}
                    previousCumulative={previousActual}
                />
            ) : (
                <SegmentRow
                    key={segment.id}
                    segmentData={segmentData}
                    split={split}
                    cumulativeTarget={targets.cumulative[segment.id]}
                    individualTarget={targets.individual[segment.id]}
                />
            );

            if (isFinalLeaf) {
                finalRow = row;
            } else {
                rows.push(row);
            }
        }

        return { rows, finalRow };
    }, [
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
    ]);
}
