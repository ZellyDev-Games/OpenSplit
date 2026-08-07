import SegmentPayload from "../../models/segmentPayload";
import { cloneSegments } from "./segmentClone";
import { findNodeMutable } from "./segmentLookup";

export function groupIntoPreviousSibling(list: SegmentPayload[], id: string): SegmentPayload[] {
    const root = cloneSegments(list);

    const found = findNodeMutable(root, id);
    if (!found) {
        return list;
    }

    const { siblings, index } = found;

    if (index <= 0) {
        return list;
    }

    const node = siblings[index];
    const previous = siblings[index - 1];

    siblings.splice(index, 1);

    previous.children = [...(previous.children ?? []), node];

    return root;
}

export function ungroupToTopLevel(list: SegmentPayload[], id: string): SegmentPayload[] {
    const root = cloneSegments(list);

    const found = findNodeMutable(root, id);
    if (!found) {
        return list;
    }

    const { siblings, index, parents } = found;

    if (parents.length === 0) {
        return list;
    }
    const node = siblings[index];

    siblings.splice(index, 1);

    const topAncestor = parents[0].node;
    const topIndex = root.findIndex((s) => s.id === topAncestor.id);
    const insertAt = topIndex >= 0 ? topIndex + 1 : root.length;

    root.splice(insertAt, 0, node);

    return root;
}
