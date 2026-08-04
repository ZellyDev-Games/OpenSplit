import { useEffect, useState } from "react";

import { FlatSegment, getAncestorIds } from "../../components/splitter/segments/segmentUtils";
import SegmentPayload from "../../models/segmentPayload";
import { log } from "../../utils/logger";

type UseExpandedParentsParams = {
    forceExpandAll?: boolean;
    currentSegmentIndex: number;
    leafSegments?: SegmentPayload[] | null;
    flatSegments: FlatSegment[];
    parentById: Map<string, string | null>;
};

export function useExpandedParents({
    forceExpandAll = false,
    currentSegmentIndex,
    leafSegments,
    flatSegments,
    parentById,
}: UseExpandedParentsParams) {
    const [expandedParents, setExpandedParents] = useState<Set<string>>(() => {
        const expanded = new Set<string>();

        flatSegments.forEach((segment) => {
            if (segment.segment.children.length > 0) {
                expanded.add(segment.segment.id);
            }
        });

        return expanded;
    });

    /*
     * Automatically expand active segment parents
     */
    useEffect(() => {
        if (forceExpandAll) {
            setExpandedParents(
                new Set(
                    flatSegments
                        .filter((segment) => segment.segment.children.length > 0)
                        .map((segment) => segment.segment.id),
                ),
            );
            return;
        }

        const leaves = leafSegments;

        if (!leaves) {
            setExpandedParents(new Set());
            return;
        }

        const active = leaves[currentSegmentIndex];

        if (!active) {
            return;
        }

        setExpandedParents(new Set(getAncestorIds(active.id, parentById)));
    }, [forceExpandAll, flatSegments, currentSegmentIndex, leafSegments, parentById]);

    useEffect(() => {
        setExpandedParents(
            new Set(
                flatSegments
                    .filter((segment) => segment.segment.children.length > 0)
                    .map((segment) => segment.segment.id),
            ),
        );
    }, [flatSegments]);

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
