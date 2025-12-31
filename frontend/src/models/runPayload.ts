import SplitPayload from "./splitPayload";
import SegmentPayload from "./segmentPayload";

export default class RunPayload {
    id: string = "";
    split_file_version: number = 0;
    total_time: number = 0;
    splits: Record<string, SplitPayload> = {};
    leaf_segments: SegmentPayload[] = [];
    completed: boolean = false;
}
