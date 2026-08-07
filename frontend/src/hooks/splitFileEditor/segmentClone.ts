import SegmentPayload from "../../models/segmentPayload";

/**
 * Deep clones the segment hierarchy.
 *
 * Mutable tree operations work on the clone so React state
 * can remain immutable.
 */
export function cloneSegments(list: SegmentPayload[]): SegmentPayload[] {
    return (list ?? []).map((seg) => {
        return new SegmentPayload({
            ...seg,
            children: cloneSegments(seg.children ?? []),
        });
    });
}
