import RunPayload from "./runPayload";
import SegmentPayload from "./segmentPayload";
import SplitFilePayload from "./splitFilePayload";

export default class SessionPayload {
    loaded_split_file: SplitFilePayload | null = null;
    leaf_segments: SegmentPayload[] | null = null;
    current_run: RunPayload | null = null;
    current_segment_index: number = -1;
    dirty: boolean = false;
}
