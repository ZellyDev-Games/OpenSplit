import RunPayload from "./runPayload";
import SegmentPayload from "./segmentPayload";

export default class SplitFilePayload {
    id: string = "";
    version: number = 1;
    game_name: string = "";
    game_category: string = "";
    window_x: number = 100
    window_y: number = 100
    window_width: number = 350
    window_height: number = 550
    segments: SegmentPayload[] = [];
    attempts: number = 0;
    runs: RunPayload[] = [];
    sob: number = 0;

    static createFrom = (source: SplitFilePayload): SplitFilePayload => {
        return { ...source };
    };
}
