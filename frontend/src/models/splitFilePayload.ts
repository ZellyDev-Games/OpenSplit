import RunPayload from "./runPayload";
import SegmentPayload from "./segmentPayload";

export default class SplitFilePayload {
    id: string = "";
    version: number = 1;
    attempts: number = 0;
    game_name: string = "";
    game_category: string = "";
    window_x: number = 100;
    window_y: number = 100;
    window_height: number = 550;
    window_width: number = 350;
    runs: RunPayload[] = [];
    segments: SegmentPayload[] = [];
    sob: number = 0;
    pb: RunPayload | null = null;
    offset: number = 0;
    autosplitter_file: string = "";

    constructor(init?: Partial<SplitFilePayload>) {
        if (init) {
            Object.assign(this, init);
        }
    }

    static createFrom = (source: Partial<SplitFilePayload>): SplitFilePayload => {
        return new SplitFilePayload(source);
    };
}
