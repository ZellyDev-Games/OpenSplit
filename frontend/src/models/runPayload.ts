import SplitPayload from "./splitPayload";

export default class RunPayload {
    id: string = "";
    split_file_version: number = 0;
    total_time: number = 0
    completed: boolean = false;
    split_payloads: SplitPayload[] = [];
}
