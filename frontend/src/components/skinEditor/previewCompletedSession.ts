import RunPayload from "../../models/runPayload";
import SplitPayload from "../../models/splitPayload";
import { createBaseSession } from "./previewBase";

export const previewSession = createBaseSession();

previewSession.current_segment_index = 8;
previewSession.dirty = false;

const run = new RunPayload();

run.id = "preview-run";
run.completed = true;
run.leaf_segments = previewSession.leaf_segments ?? [];

run.splits = {};

const previewSplits = [
    { id: "opening", duration: 12000, cumulative: 12000 },
    { id: "world1-level1", duration: 30000, cumulative: 42000 },
    { id: "world1-level2", duration: 46100, cumulative: 82800 },
    { id: "world1-boss", duration: 18700, cumulative: 101500 },
    { id: "castle-entry", duration: 15100, cumulative: 116600 },
    { id: "castle-tower", duration: 61400, cumulative: 178000 },
    { id: "final-boss", duration: 96200, cumulative: 274200 },
    { id: "credits", duration: 25000, cumulative: 299200 },
];

let cumulative = 0;

for (const { id, duration } of previewSplits) {
    cumulative += duration;

    run.splits[id] = new SplitPayload({
        split_segment_id: id,
        current_duration: duration,
        current_cumulative: cumulative,
    });
}

run.total_time = cumulative;

previewSession.current_run = run;
