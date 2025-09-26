import RunPayload from "./runPayload";
import SegmentPayload from "./segmentPayload";

export default class SplitFilePayload {
    id: string = "";
    version: number = 1;
    game_name: string = "";
    game_category: string = "";
    window_x: number = 100
    window_y: number = 100
    window_height: number = 550
    window_width: number = 350
    runs: RunPayload[] = [];
    segments: SegmentPayload[] = [];
    sob: number = 0;
    pb: RunPayload | null = null;

    static createFrom = (source: SplitFilePayload): SplitFilePayload => {
        return { ...source };
    };
}
