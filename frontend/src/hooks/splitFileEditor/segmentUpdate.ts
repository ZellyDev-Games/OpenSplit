import SegmentPayload from "../../models/segmentPayload";

export function updateSegmentRecursive(
    segments: SegmentPayload[],
    id: string,
    updater: (segment: SegmentPayload) => SegmentPayload,
): SegmentPayload[] {
    return segments.map((segment) => {
        if (segment.id === id) {
            return updater(segment);
        }

        return new SegmentPayload({
            ...segment,

            children: updateSegmentRecursive(segment.children ?? [], id, updater),
        });
    });
}
