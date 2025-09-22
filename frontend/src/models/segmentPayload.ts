import StatTime from "./statTime";

export default class SegmentPayload {
    id: string = "";
    name: string = "";
    gold: StatTime = new StatTime();
    average: StatTime = new StatTime();
    pb: StatTime = new StatTime();

    static createFrom = (source: SegmentPayload): SegmentPayload => {
        return { ...source };
    };
}
