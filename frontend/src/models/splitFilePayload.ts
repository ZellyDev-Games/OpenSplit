import RunPayload from "./runPayload";
import SegmentPayload from "./segmentPayload";
import WorldRecord from "./worldRecord";

export default class SplitFilePayload {
    id: string = "";
    game_name: string = "";
    speedrun_game_id = "";
    game_category: string = "";
    speedrun_game_category_id = "";
    version: number = 1;

    selected_skin?: string;

    segments: SegmentPayload[] = [];
    runs: RunPayload[] = [];
    pb: RunPayload | null = null;

    sob: number = 0;
    attempts: number = 0;
    offset: number = 0;
    platform: string = "SNES";

    wr: WorldRecord = new WorldRecord();

    window_x: number = 100;
    window_y: number = 100;
    window_height: number = 550;
    window_width: number = 350;

    constructor(init?: Partial<SplitFilePayload>) {
        if (init) {
            Object.assign(this, init);
        }
    }

    static createFrom = (source: Partial<SplitFilePayload>): SplitFilePayload => {
        return new SplitFilePayload(source);
    };
}
