/**
 * Complete split file loaded from disk.
 *
 * Contains:
 *  - segment tree
 *  - historical runs
 *  - PB
 *  - WR metadata
 *  - window state
 *  - skin configuration
 */

import RunPayload from "./runPayload";
import SegmentPayload from "./segmentPayload";
import WorldRecord from "./worldRecord";

export type SplitterLayout = "vertical" | "horizontal";

export type SplitterWindow = {
    x: number;
    y: number;
    width: number;
    height: number;
};

export type SplitterWindows = {
    vertical: SplitterWindow;
    horizontal: SplitterWindow;
};

export type SplitVariable = {
    id: string;
    name: string;
    value: string;
    label: string;
};

export default class SplitFilePayload {
    id: string = "";
    game_name: string = "";
    speedrun_game_id = "";
    game_category: string = "";
    speedrun_game_category_id = "";
    variables: SplitVariable[] = [];

    version: number = 0;

    selected_skin?: string;

    segments: SegmentPayload[] = [];
    runs: RunPayload[] = [];
    pb: RunPayload | null = null;

    sob: number = 0;
    attempts: number = 0;
    offset: number = 0;
    platform: string = "SNES";

    wr: WorldRecord = new WorldRecord();

    /**
     * Persisted splitter layout.
     *
     * An empty value means that the split file has not explicitly
     * selected a layout and the skin's --splitter-layout value should
     * be used as the default.
     */
    layout: SplitterLayout | "" = "";

    /**
     * Persisted window geometry for each splitter layout.
     *
     * Each layout maintains its own independent position and size.
     */
    windows: SplitterWindows = {
        vertical: {
            x: 100,
            y: 100,
            width: 350,
            height: 550,
        },
        horizontal: {
            x: 100,
            y: 100,
            width: 900,
            height: 400,
        },
    };

    constructor(init?: Partial<SplitFilePayload>) {
        if (init) {
            Object.assign(this, init);
        }
    }

    static createFrom = (source: Partial<SplitFilePayload>): SplitFilePayload => {
        return new SplitFilePayload(source);
    };
}
