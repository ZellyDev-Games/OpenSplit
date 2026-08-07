import SegmentPayload from "../../models/segmentPayload";

type ParentRef = {
    node: SegmentPayload;
    siblings: SegmentPayload[];
    index: number;
};

/**
 * Finds a node and returns mutable references to:
 *   - the sibling array containing the node
 *   - the node index
 *   - every ancestor
 *
 * Used by move/group/ungroup operations.
 */
export function findNodeMutable(
    siblings: SegmentPayload[],
    id: string,
    parents: ParentRef[] = [],
): { siblings: SegmentPayload[]; index: number; parents: ParentRef[] } | null {
    for (let i = 0; i < siblings.length; i++) {
        const node = siblings[i];

        if (node.id === id) {
            return { siblings, index: i, parents };
        }

        const kids = node.children ?? [];

        if (kids.length > 0) {
            const nextParents = parents.concat([{ node, siblings, index: i }]);
            const found = findNodeMutable(kids, id, nextParents);

            if (found) {
                return found;
            }
        }
    }

    return null;
}
