import SplitPayload from "./splitPayload";
import StatTime from "./statTime";

export default class RunPayload {
    id: string = "";
    split_file_version: number = 0;
    total_time: StatTime = new StatTime();
    completed: boolean = false;
    split_payloads: SplitPayload[] = [];
}
