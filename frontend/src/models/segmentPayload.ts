export default class SegmentPayload {
    id: string = "";
    name: string = "";
    gold: number = 0;
    average: number = 0;
    pb: number = 0;

    static createFrom = (source: SegmentPayload): SegmentPayload => {
        return { ...source };
    };
}
