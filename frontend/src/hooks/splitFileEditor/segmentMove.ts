import SegmentPayload from "../../models/segmentPayload";
import { cloneSegments } from "./segmentClone";
import { findNodeMutable } from "./segmentLookup";

export function moveSegmentUp(list: SegmentPayload[], id: string): SegmentPayload[] {
    const root = cloneSegments(list);

    const found = findNodeMutable(root, id);
    if (!found) return list;

    const { siblings, index } = found;

    if (index <= 0) {
        return list;
    }

    [siblings[index - 1], siblings[index]] = [siblings[index], siblings[index - 1]];

    return root;
}

export function moveSegmentDown(list: SegmentPayload[], id: string): SegmentPayload[] {
    const root = cloneSegments(list);

    const found = findNodeMutable(root, id);
    if (!found) return list;

    const { siblings, index } = found;

    if (index >= siblings.length - 1) {
        return list;
    }

    [siblings[index], siblings[index + 1]] = [siblings[index + 1], siblings[index]];

    return root;
}
