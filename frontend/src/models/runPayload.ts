import SplitPayload from "./splitPayload";

export default class RunPayload {
    id: string = "";
    split_file_id: string = "";
    split_file_version: number = 0;
    total_time: number = 0;
    splits: SplitPayload[] = [];
    completed: boolean = false;
}
