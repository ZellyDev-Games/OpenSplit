export default class SegmentPayload {
    id: string;
    name: string = "";
    gold: number = 0;
    average: number = 0;
    pb: number = 0;
    parent: string | null = null;
    children: SegmentPayload[] = [];

    constructor() {
        this.id = crypto.randomUUID();
    }

    static createFrom(obj: any): SegmentPayload {
        const s = new SegmentPayload();

        // ONLY override ID if the source actually has a meaningful ID
        if (obj.id && typeof obj.id === "string" && obj.id.trim() !== "") {
            s.id = obj.id;
        }

        s.name = obj.name ?? "";
        s.average = obj.average ?? null;
        s.pb = obj.pb ?? null;
        s.gold = obj.gold ?? null;
        s.parent = obj.parent ?? null;

        s.children = (obj.children ?? []).map((c: any) => SegmentPayload.createFrom(c));

        return s;
    }
}
