import StatTime from "./statTime";

export default class SplitPayload {
    split_index: number = 0;
    split_segment_id: string = "";
    current_time: StatTime = new StatTime();
}
