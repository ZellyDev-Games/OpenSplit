export default class SegmentPayload {
    id: string;
    name: string = "";
    gold: number = 0;
    average: number = 0;
    pb: number = 0;
    parent: string | null = null;
    children: SegmentPayload[] = [];

    constructor(init?: Partial<SegmentPayload>) {
        this.id = init?.id ?? crypto.randomUUID();
        this.name = init?.name ?? "";
        this.gold = init?.gold ?? 0;
        this.average = init?.average ?? 0;
        this.pb = init?.pb ?? 0;
        this.parent = init?.parent ?? null;
        this.children = (init?.children ?? []).map((c) => new SegmentPayload(c));
    }
}

export class FlattenedSegmentPayload {
    segment: SegmentPayload;
    depth: number;

    constructor(init: SegmentPayload, depth: number = 0) {
        this.segment = init;
        this.depth = depth;
    }
}
