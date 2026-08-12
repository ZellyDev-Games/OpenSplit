/**
 * Represents the current runtime state of OpenSplit.
 *
 * This object changes as the runner progresses through
 * a split file.
 *
 * Persisted splitter window geometry is stored on the loaded
 * split file rather than on the runtime session.
 */

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
