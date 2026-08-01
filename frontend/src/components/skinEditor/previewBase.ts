import { Command } from "../../models/command";
import type { ConfigPayload } from "../../models/configPayload";
import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";
import SplitFilePayload from "../../models/splitFilePayload";
import WorldRecord from "../../models/worldRecord";

export function createBaseSession(): SessionPayload {
    const session = new SessionPayload();

    previewSplitFile.segments = previewSegments;

    session.loaded_split_file = previewSplitFile;
    session.leaf_segments = getLeafSegments(previewSegments);
    session.dirty = false;

    return session;
}

// Preview split hierarchy
const previewSegments = [
    new SegmentPayload({
        id: "opening",
        name: "Opening",
        gold: 12500,
        average: 14200,
        pb: 13800,
    }),

    new SegmentPayload({
        id: "world1",
        name: "World 1",
        children: [
            new SegmentPayload({
                id: "world1-level1",
                name: "Level 1",
                gold: 28400,
                average: 31200,
                pb: 30500,
            }),
            new SegmentPayload({
                id: "world1-level2",
                name: "Level 2",
                gold: 42000,
                average: 44800,
                pb: 46100,
            }),
            new SegmentPayload({
                id: "world1-boss",
                name: "World 1 Boss",
                gold: 18700,
                average: 20500,
                pb: 19900,
            }),
        ],
    }),

    new SegmentPayload({
        id: "castle",
        name: "Castle",
        children: [
            new SegmentPayload({
                id: "castle-entry",
                name: "Entry",
                gold: 15100,
                average: 16800,
                pb: 17400,
            }),
            new SegmentPayload({
                id: "castle-tower",
                name: "Tower",
                gold: 55400,
                average: 58200,
                pb: 54900,
            }),
        ],
    }),

    new SegmentPayload({
        id: "final-boss",
        name: "Final Boss",
        gold: 96200,
        average: 102400,
        pb: 108600,
    }),

    new SegmentPayload({
        id: "credits",
        name: "Credits",
        gold: 25000,
        average: 27500,
        pb: 26800,
    }),
];

function getLeafSegments(segments: SegmentPayload[]): SegmentPayload[] {
    const leaves: SegmentPayload[] = [];

    for (const segment of segments) {
        if (segment.children.length === 0) {
            leaves.push(segment);
            continue;
        }

        leaves.push(...getLeafSegments(segment.children));
    }

    return leaves;
}

const previewWorldRecord = new WorldRecord();

previewWorldRecord.show = true;
previewWorldRecord.run_id = "preview-wr";
previewWorldRecord.players = ["Zoast"];
previewWorldRecord.real_time = 2487.31;
previewWorldRecord.in_game_time = 2459.42;

export const previewSplitFile = new SplitFilePayload({
    id: "preview",

    game_name: "Super Metroid",
    game_category: "Any%",

    speedrun_game_id: "",
    speedrun_game_category_id: "",

    attempts: 324,
    offset: 0,

    variables: [
        {
            id: "difficulty",
            name: "difficulty",
            label: "Difficulty",
            value: "Normal",
        },
        {
            id: "platform",
            name: "platform",
            label: "Platform",
            value: "SNES",
        },
    ],

    segments: [],

    wr: previewWorldRecord,

    window_width: 320,
    window_height: 580,
    window_x: 0,
    window_y: 0,

    platform: "SNES",
});

export const previewConfig: ConfigPayload = {
    speed_run_API_base: "",

    key_config: {} as Record<Command, never>,

    global_hotkeys_active: false,

    selected_skin: "default",

    rolling_average_runs: 10,
};

previewSplitFile.segments = previewSegments;
