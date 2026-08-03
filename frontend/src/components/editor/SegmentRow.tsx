/**
 * Renders a single segment row within the split editor.
 *
 * Supports nested segment groups, editing timing values,
 * grouping operations, icons, and child segment creation.
 */

import {
    faArrowDown,
    faArrowRightFromBracket,
    faArrowUp,
    faArrowUpFromBracket,
    faFolder,
    faTrash,
} from "@fortawesome/free-solid-svg-icons";
import React from "react";

import SegmentPayload from "../../models/segmentPayload";
import { colorFromId, GroupCtx } from "./hashColor";
import SegmentIconPicker from "./SegmentIconPicker";
import { TimeRow } from "./TimeRow";
import { IconButton } from "./Tooltip";

type SegmentRowProps = {
    segment: SegmentPayload;

    depth: number;
    index: number;

    displayAverage: number;
    displayPB: number;
    displayGold: number;

    inheritedGroup: GroupCtx | null;
    isDirectChild: boolean;

    onMoveUp(id: string): void;
    onMoveDown(id: string): void;
    onGroup(id: string): void;
    onUngroup(id: string): void;
    onDelete(id: string): void;
    onAddChild(parent: SegmentPayload): void;

    onUpdate(id: string, updater: (segment: SegmentPayload) => SegmentPayload): void;
};

export default function SegmentRow({
    segment,
    depth,
    index,
    inheritedGroup,
    isDirectChild,
    displayAverage,
    displayPB,
    displayGold,
    onMoveUp,
    onMoveDown,
    onGroup,
    onUngroup,
    onDelete,
    onAddChild,
    onUpdate,
}: SegmentRowProps) {
    const hasChildren = (segment.children ?? []).length > 0;

    const ownGroup = hasChildren ? { bg: colorFromId(segment.id) } : null;

    // Group parent gets its own color.
    // Direct children inherit the parent's color.
    const rowGroup = ownGroup ?? (isDirectChild ? inheritedGroup : null);

    const rowStyle = rowGroup
        ? ({
              ["--group-bg"]: rowGroup.bg,
          } as React.CSSProperties)
        : undefined;

    const rowClassName = [
        rowGroup && "seg-group",
        ownGroup && "seg-group-parent",
        !ownGroup && isDirectChild && inheritedGroup && "seg-group-child",
    ]
        .filter(Boolean)
        .join(" ");

    return (
        <>
            <tr className={rowClassName} style={rowStyle}>
                {/* move buttons */}
                <td>
                    <div
                        style={{
                            display: "flex",
                            flexDirection: "row",
                            alignItems: "center",
                            gap: 6,
                        }}
                    >
                        <IconButton icon={faArrowUp} tooltip="Move segment up" onClick={() => onMoveUp(segment.id)} />

                        <IconButton
                            icon={faArrowDown}
                            tooltip="Move segment down"
                            onClick={() => onMoveDown(segment.id)}
                        />

                        <IconButton
                            icon={faArrowUpFromBracket}
                            tooltip="Group under previous segment"
                            show={index !== 0}
                            onClick={() => onGroup(segment.id)}
                        />

                        <IconButton
                            icon={faArrowRightFromBracket}
                            tooltip="Remove from group"
                            show={depth > 0}
                            onClick={() => onUngroup(segment.id)}
                        />
                    </div>
                </td>

                {/* icon picker */}
                <td>
                    <SegmentIconPicker
                        icon={segment.icon}
                        onChange={(icon) =>
                            onUpdate(segment.id, (s) => ({
                                ...s,
                                icon,
                            }))
                        }
                    />
                </td>

                {/* name */}
                <td style={{ paddingLeft: depth * 20 }}>
                    <input
                        value={segment.name}
                        onChange={(e) =>
                            onUpdate(segment.id, (s) => ({
                                ...s,
                                name: e.target.value,
                            }))
                        }
                    />
                </td>

                {/* average */}
                <td>
                    {!hasChildren && (
                        <TimeRow
                            time={displayAverage}
                            onChange={(ms) =>
                                onUpdate(segment.id, (s) => ({
                                    ...s,
                                    average: ms,
                                }))
                            }
                        />
                    )}
                </td>

                {/* pb */}
                <td>
                    {!hasChildren && (
                        <TimeRow
                            time={displayPB}
                            onChange={(ms) =>
                                onUpdate(segment.id, (s) => ({
                                    ...s,
                                    pb: ms,
                                }))
                            }
                        />
                    )}
                </td>

                {/* gold */}
                <td>
                    {!hasChildren && (
                        <TimeRow
                            time={displayGold}
                            onChange={(ms) =>
                                onUpdate(segment.id, (s) => ({
                                    ...s,
                                    gold: ms,
                                }))
                            }
                        />
                    )}
                </td>

                {/* add child */}
                <td>
                    <IconButton icon={faFolder} tooltip="Add subsegment" onClick={() => onAddChild(segment)} />
                </td>

                {/* delete */}
                <td>
                    <IconButton icon={faTrash} tooltip="Delete segment" onClick={() => onDelete(segment.id)} />
                </td>
            </tr>
        </>
    );
}
