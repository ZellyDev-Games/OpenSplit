import SplitFilePayload from "./splitFilePayload";
import SegmentPayload from "./segmentPayload";
import StatTime from "./statTime";
import RunPayload from "./runPayload";

export default class SessionPayload {
    split_file: SplitFilePayload | null = null;
    current_segment_index: number = -1;
    current_segment: SegmentPayload | null = null;
    finished: boolean = false;
    paused: boolean = false;
    current_time: StatTime = new StatTime();
    current_run: RunPayload | null = null;
}
