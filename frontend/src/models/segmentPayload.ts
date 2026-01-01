export default class SegmentPayload {
    id: string;
    name: string = "";
    gold: number = 0;
    average: number = 0;
    pb: number = 0;
    children: SegmentPayload[] = [];

    constructor(init?: Partial<SegmentPayload>) {
        this.id = init?.id ?? crypto.randomUUID();
        this.name = init?.name ?? "";
        this.gold = init?.gold ?? 0;
        this.average = init?.average ?? 0;
        this.pb = init?.pb ?? 0;
        this.children = (init?.children ?? []).map((c) => new SegmentPayload(c));
    }
}
