import SegmentPayload from "../../../models/segmentPayload";

export type SegmentUpdater = (id: string, updater: (segment: SegmentPayload) => SegmentPayload) => void;
