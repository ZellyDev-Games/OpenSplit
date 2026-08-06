import { useEffect, useState } from "react";

import { FlatSegment, getAncestorIds } from "../../components/splitter/segments/segmentUtils";
import SegmentPayload from "../../models/segmentPayload";
import { log } from "../../utils/logger";

type UseExpandedParentsParams = {
    forceExpandAll?: boolean;
    currentSegmentIndex: number;
    runActive: boolean;
    leafSegments?: SegmentPayload[] | null;
    flatSegments: FlatSegment[];
    parentById: Map<string, string | null>;
};

export function useExpandedParents({
    forceExpandAll = false,
    runActive,
    currentSegmentIndex,
    leafSegments,
    flatSegments,
    parentById,
}: UseExpandedParentsParams) {
    const expandAll = () =>
        new Set(
            flatSegments.filter((segment) => segment.segment.children.length > 0).map((segment) => segment.segment.id),
        );

    const [expandedParents, setExpandedParents] = useState<Set<string>>(() => expandAll());

    useEffect(() => {
        if (forceExpandAll || !runActive) {
            setExpandedParents(expandAll());
            return;
        }

        const active = leafSegments?.[currentSegmentIndex];

        if (!active) {
            setExpandedParents(expandAll());
            return;
        }

        // Running: collapse everything except active branch
        setExpandedParents(new Set(getAncestorIds(active.id, parentById)));
    }, [forceExpandAll, runActive, currentSegmentIndex, leafSegments, parentById, flatSegments]);

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

    return {
        expandedParents,
        toggleParent,
    };
}
