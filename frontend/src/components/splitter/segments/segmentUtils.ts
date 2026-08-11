import SegmentPayload from "../../../models/segmentPayload";

/**
 * Flattened representation of the segment tree.
 * Used for rendering while preserving hierarchy information.
 */
export type FlatSegment = {
    segment: SegmentPayload;
    depth: number;
    parentId: string | null;
    hasChildren: boolean;
};

/**
 * Cached comparison targets for every leaf segment.
 *
 * cumulative:
 *     Total expected time up to this split.
 *
 * individual:
 *     Expected duration of this split alone.
 */
export type Targets = {
    cumulative: Record<string, number>;
    individual: Record<string, number>;
};

/**
 * Converts the recursive tree into a depth-first list.
 */
export function flattenSegments(
    segments: SegmentPayload[],
    depth: number = 0,
    parentId: string | null = null,
): FlatSegment[] {
    const flat: FlatSegment[] = [];

    for (const segment of segments) {
        flat.push({
            segment: segment,
            depth: depth,
            parentId: parentId,
            hasChildren: segment.children.length > 0,
        });

        if (segment.children.length > 0) {
            flat.push(...flattenSegments(segment.children, depth + 1, segment.id));
        }
    }

    return flat;
}

export function getAncestorIds(leafId: string, parentById: Map<string, string | null>): string[] {
    const ancestors: string[] = [];
    let current: string | null | undefined = leafId;

    while (current != null) {
        const parent = parentById.get(current);

        if (parent == null) {
            break;
        }

        ancestors.push(parent);
        current = parent;
    }

    return ancestors;
}

export function isVisible(id: string, parentById: Map<string, string | null>, expandedParents: Set<string>): boolean {
    let current: string | null | undefined = id;

    while (current != null) {
        const parent = parentById.get(current);

        if (parent == null) {
            return true;
        }

        if (!expandedParents.has(parent)) {
            return false;
        }

        current = parent;
    }

    return true;
}

/**
 * Returns whether an element is completely visible inside a container.
 *
 * Checks both axes so this works for:
 *
 *  - vertical splitter layout
 *  - horizontal splitter layout
 *
 * An element is considered visible only when its complete bounding
 * rectangle is inside the container's visible rectangle.
 */
export function isElementFullyVisible(element: HTMLElement, container: HTMLElement): boolean {
    const elementRect = element.getBoundingClientRect();
    const containerRect = container.getBoundingClientRect();

    return (
        elementRect.top >= containerRect.top &&
        elementRect.bottom <= containerRect.bottom &&
        elementRect.left >= containerRect.left &&
        elementRect.right <= containerRect.right
    );
}
