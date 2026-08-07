import {
    faArrowDown,
    faArrowRightFromBracket,
    faArrowUp,
    faArrowUpFromBracket,
} from "@fortawesome/free-solid-svg-icons";

import SegmentPayload from "../../../models/segmentPayload";
import { IconButton } from "../shared/IconButton";

type SegmentActionsProps = {
    segment: SegmentPayload;

    index: number;

    depth: number;

    onMoveUp(id: string): void;

    onMoveDown(id: string): void;

    onGroup(id: string): void;

    onUngroup(id: string): void;
};

export default function SegmentActions({
    segment,

    index,

    depth,

    onMoveUp,

    onMoveDown,

    onGroup,

    onUngroup,
}: SegmentActionsProps) {
    return (
        <div className="segment-actions">
            <div className="segment-action-up">
                <IconButton icon={faArrowUp} tooltip="Move segment up" onClick={() => onMoveUp(segment.id)} />
            </div>

            <div className="segment-action-down">
                <IconButton icon={faArrowDown} tooltip="Move segment down" onClick={() => onMoveDown(segment.id)} />
            </div>

            <div className="segment-action-group">
                <IconButton
                    icon={faArrowUpFromBracket}
                    tooltip="Group under previous segment"
                    show={index !== 0}
                    onClick={() => onGroup(segment.id)}
                />
            </div>

            <div className="segment-action-ungroup">
                <IconButton
                    icon={faArrowRightFromBracket}
                    tooltip="Remove from group"
                    show={depth > 0}
                    onClick={() => onUngroup(segment.id)}
                />
            </div>
        </div>
    );
}

/**
 * Adds a new child segment beneath the specified parent.
 *
 * Returns a new immutable tree.
 */
export function addChildRecursive(list: SegmentPayload[], parent: SegmentPayload): SegmentPayload[] {
    return list.map((item) => {
        if (item.id === parent.id) {
            const child = new SegmentPayload();

            return {
                ...item,
                children: [...(item.children ?? []), child],
            };
        }

        return {
            ...item,
            children: addChildRecursive(item.children ?? [], parent),
        };
    });
}

export function deleteSegmentRecursive(list: SegmentPayload[], id: string): SegmentPayload[] {
    return list
        .filter((seg) => seg.id !== id)
        .map((seg) => ({
            ...seg,
            children: deleteSegmentRecursive(seg.children ?? [], id),
        }));
}
