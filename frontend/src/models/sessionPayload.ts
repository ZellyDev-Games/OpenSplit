import RunPayload from "./runPayload";
import SplitFilePayload from "./splitFilePayload";
import {State} from "../App"

export default class SessionPayload {
    loaded_split_file: SplitFilePayload | null = null;
    current_run: RunPayload | null = null;
    current_segment_index: number = -1;
    session_state: State = State.WELCOME;
    dirty: boolean = false
}
