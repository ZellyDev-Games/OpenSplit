import RunPayload from "./runPayload";
import StatTime from "./statTime";
import WindowParams from "./windowParams";
import SegmentPayload from "./segmentPayload";

export default class SplitFilePayload {
    id: string = "";
    version: number = 1;
    game_name: string = "";
    game_category: string = "";
    segments: SegmentPayload[] = [];
    attempts: number = 0;
    runs: RunPayload[] = [];
    sob: StatTime = new StatTime();
    window_params: WindowParams = new WindowParams();

    static createFrom = (source : SplitFilePayload): SplitFilePayload => {
        return {...source}
    }
}
