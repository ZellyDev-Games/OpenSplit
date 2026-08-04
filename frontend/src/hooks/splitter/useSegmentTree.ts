import { useMemo } from "react";

import { FlatSegment, flattenSegments, getAncestorIds } from "../../components/splitter/segments/segmentUtils";
import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";

export function useSegmentTree(sessionPayload: SessionPayload) {
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
            map.set(segment.segment.id, segment.parentId);
        });

        return map;
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

    return {
        flatSegments,
        leafIndexById,
        parentById,
        lastLeafByParentId,
        leavesByParentId,
        finalLeafId,
    };
}
