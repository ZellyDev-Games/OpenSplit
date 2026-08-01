/**
 * Timing information for a single completed segment
 * within a run.
 */

export default class SplitPayload {
    split_segment_id: string = "";
    current_cumulative: number = 0;
    current_duration: number = 0;

    constructor(init?: Partial<SplitPayload>) {
        Object.assign(this, init);
    }
}
