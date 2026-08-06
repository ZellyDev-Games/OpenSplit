/**
 * Renders a single segment row within the split editor.
 *
 * Supports nested segment groups, editing timing values,
 * grouping operations, icons, and child segment creation.
 */

import { CSSProperties } from "react";

import SegmentPayload from "../../../models/segmentPayload";
import SegmentActionCell from "./cells/SegmentActionCell";
import SegmentChildActionsCell from "./cells/SegmentChildActionsCell";
import SegmentIconCell from "./cells/SegmentIconCell";
import SegmentNameCell from "./cells/SegmentNameCell";
import SegmentTimeCells from "./cells/SegmentTimeCells";
import { colorFromId, GroupCtx } from "./hashColor";

type SegmentRowProps = {
    segment: SegmentPayload;

    depth: number;

    index: number;

    average: number;

    pb: number;

    gold: number;

    inheritedGroup: GroupCtx | null;

    isDirectChild: boolean;

    onMoveUp(id: string): void;

    onMoveDown(id: string): void;

    onGroup(id: string): void;

    onUngroup(id: string): void;

    onDelete(id: string): void;

    onAddChild(segment: SegmentPayload): void;

    onUpdate(id: string, updater: (segment: SegmentPayload) => SegmentPayload): void;
};

export default function SegmentRow({
    segment,
    depth,
    index,
    average,
    pb,
    gold,
    inheritedGroup,
    isDirectChild,
    onMoveUp,
    onMoveDown,
    onGroup,
    onUngroup,
    onDelete,
    onAddChild,
    onUpdate,
}: SegmentRowProps) {
    const hasChildren = (segment.children ?? []).length > 0;

    const ownGroup = hasChildren
        ? {
              bg: colorFromId(segment.id),
          }
        : null;

    const rowGroup = ownGroup ?? (isDirectChild ? inheritedGroup : null);

    const rowStyle = rowGroup
        ? ({
              "--group-bg": rowGroup.bg,
          } as CSSProperties)
        : undefined;

    const rowClassName = [
        rowGroup && "seg-group",
        ownGroup && "seg-group-parent",
        !ownGroup && isDirectChild && inheritedGroup && "seg-group-child",
    ]
        .filter(Boolean)
        .join(" ");

    return (
        <tr className={rowClassName} style={rowStyle}>
            <SegmentActionCell
                segment={segment}
                index={index}
                depth={depth}
                onMoveUp={onMoveUp}
                onMoveDown={onMoveDown}
                onGroup={onGroup}
                onUngroup={onUngroup}
            />

            <SegmentIconCell segment={segment} onUpdate={onUpdate} />

            <SegmentNameCell segment={segment} depth={depth} onUpdate={onUpdate} />

            <SegmentTimeCells segment={segment} average={average} pb={pb} gold={gold} onUpdate={onUpdate} />

            <SegmentChildActionsCell segment={segment} onAddChild={onAddChild} onDelete={onDelete} />
        </tr>
    );
}
