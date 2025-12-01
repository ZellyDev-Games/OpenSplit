interface SegmentJSON {
    id?: string;
    name?: string;
    average?: number;
    pb?: number;
    gold?: number;
    parent?: string;
    children?: SegmentJSON[];
}

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

    static createFrom(obj: SegmentJSON): SegmentPayload {
        const s = new SegmentPayload();

        // ONLY override ID if the source actually has a meaningful ID
        if (typeof obj.id === "string" && obj.id.trim() !== "") {
            s.id = obj.id;
        }

        if (typeof obj.name === "string") {
            s.name = obj.name;
        }

        if (typeof obj.average === "number") {
            s.average = obj.average;
        }

        if (typeof obj.pb === "number") {
            s.pb = obj.pb;
        }

        if (typeof obj.gold === "number") {
            s.gold = obj.gold;
        }

        if (typeof obj.parent === "string") {
            s.parent = obj.parent;
        }

        // children: must be an array to recurse
        if (Array.isArray(obj.children)) {
            s.children = obj.children.map((c) => SegmentPayload.createFrom(c as SegmentJSON));
        } else {
            s.children = [];
        }

        return s;
    }
}
