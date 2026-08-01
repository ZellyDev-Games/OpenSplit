/**
 * One segment in the split hierarchy.
 *
 * Parent segments contain children.
 * Leaf segments are individually timed.
 */

export default class SegmentPayload {
    id: string;
    name: string = "";
    gold: number = -1;
    average: number = -1;
    pb: number = -1;
    icon = "";
    children: SegmentPayload[] = [];

    constructor(init?: Partial<SegmentPayload>) {
        this.id = init?.id ?? crypto.randomUUID();
        this.name = init?.name ?? "";
        this.gold = init?.gold ?? -1;
        this.average = init?.average ?? -1;
        this.pb = init?.pb ?? -1;
        this.icon = init?.icon ?? "";
        this.children = (init?.children ?? []).map((c) => new SegmentPayload(c));
    }
}
