import RunPayload from "../../../../models/runPayload";
import SegmentPayload from "../../../../models/segmentPayload";
import SplitPayload from "../../../../models/splitPayload";

interface PreviewSplit {
    id: string;
    duration: number;
}

export function createPreviewRun(
    completed: boolean,
    splits: PreviewSplit[],
    leafSegments: SegmentPayload[],
): RunPayload {
    const run = new RunPayload();

    run.id = "preview-run";

    run.completed = completed;

    run.leaf_segments = leafSegments;

    run.splits = {};

    let cumulative = 0;

    for (const split of splits) {
        cumulative += split.duration;

        run.splits[split.id] = new SplitPayload({
            split_segment_id: split.id,
            current_duration: split.duration,
            current_cumulative: cumulative,
        });
    }

    run.total_time = cumulative;

    return run;
}
