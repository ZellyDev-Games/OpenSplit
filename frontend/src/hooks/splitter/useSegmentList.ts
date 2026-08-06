import SessionPayload from "../../models/sessionPayload";
import { Comparison } from "./useComparison";
import { useComparisonTargets } from "./useComparisonTargets";
import { useExpandedParents } from "./useExpandedParents";
import { useSegmentCompletion } from "./useSegmentCompletion";
import { useSegmentRows } from "./useSegmentRows";
import { useSegmentScroll } from "./useSegmentScroll";
import { useSegmentTree } from "./useSegmentTree";

type UseSegmentListParams = {
    sessionPayload: SessionPayload;
    comparison: Comparison;
    forceExpandAll?: boolean;
};

export function useSegmentList({ sessionPayload, comparison, forceExpandAll = false }: UseSegmentListParams) {
    const completeClassName = useSegmentCompletion(sessionPayload);

    const targets = useComparisonTargets(comparison, sessionPayload.leaf_segments);

    const tree = useSegmentTree(sessionPayload);

    const runActive = sessionPayload.current_run !== null;

    const { expandedParents, toggleParent } = useExpandedParents({
        forceExpandAll,
        runActive,
        currentSegmentIndex: sessionPayload.current_segment_index,
        leafSegments: sessionPayload.leaf_segments,
        flatSegments: tree.flatSegments,
        parentById: tree.parentById,
    });

    const { containerRef, activeRowRef } = useSegmentScroll(sessionPayload.current_segment_index);

    const { rows, finalRow } = useSegmentRows({
        sessionPayload,
        targets,
        completeClassName,
        expandedParents,
        toggleParent,
        activeRowRef,
        ...tree,
    });

    return {
        completeClassName,
        rows,
        finalRow,
        containerRef,
        activeRowRef,
    };
}
