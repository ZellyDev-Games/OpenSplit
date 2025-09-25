import RunPayload from "./runPayload";
import SplitFilePayload from "./splitFilePayload";

export default class SessionPayload {
    loaded_split_file: SplitFilePayload | null = null;
    current_segment_index: number = -1;
    finished: boolean = false;
    paused: boolean = false;
    current_time: number = 0
    dirty: boolean = false
    current_run: RunPayload | null = null;
    session_state: number = 0;
}
