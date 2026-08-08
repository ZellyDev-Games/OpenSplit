import { Dispatch, ReactElement, SetStateAction } from "react";

import { groupIntoPreviousSibling, ungroupToTopLevel } from "../../../hooks/splitFileEditor/segmentGroup";
import { moveSegmentDown, moveSegmentUp } from "../../../hooks/splitFileEditor/segmentMove";
import { addTime } from "../../../hooks/splitFileEditor/useSegmentRows";
import SegmentPayload from "../../../models/segmentPayload";
import { RenderResult, RunningTotals } from "../types/render";
import { SegmentUpdater } from "../types/segment";
import { colorFromId, GroupCtx } from "./hashColor";
import SegmentRow from "./SegmentRow";

type RenderSegmentRowsProps = {
    segments: SegmentPayload[];
    setSegments?: Dispatch<SetStateAction<SegmentPayload[]>>;
    depth?: number;
    inheritedGroup?: GroupCtx | null;
    isDirectChild?: boolean;

    totals?: RunningTotals;

    showCumulativeTimes: boolean;

    onDelete: (id: string) => void;
    onAddChild: (parent: SegmentPayload | null) => void;
    onUpdate: SegmentUpdater;
};

export function renderSegmentRows({
    segments,
    setSegments,
    depth = 0,
    inheritedGroup = null,
    isDirectChild = false,
    totals = {
        avg: 0,
        pb: 0,
        gold: 0,
    },

    showCumulativeTimes,

    onDelete,
    onAddChild,
    onUpdate,
}: RenderSegmentRowsProps): RenderResult {
    let running = { ...totals };
    const rows: ReactElement[] = [];

    const updateTree = (fn: (segments: SegmentPayload[]) => SegmentPayload[]) => {
        setSegments?.(fn);
    };

    for (let i = 0; i < segments.length; i++) {
        const segment = segments[i];

        const children = segment.children ?? [];

        const hasChildren = children.length > 0;

        const displayAverage =
            segment.average < 0 ? -1 : showCumulativeTimes ? running.avg + segment.average : segment.average;

        const displayPB = segment.pb < 0 ? -1 : showCumulativeTimes ? running.pb + segment.pb : segment.pb;

        const displayGold = segment.gold < 0 ? -1 : showCumulativeTimes ? running.gold + segment.gold : segment.gold;

        const childResult = hasChildren
            ? renderSegmentRows({
                  segments: children,
                  depth: depth + 1,
                  inheritedGroup: {
                      bg: colorFromId(segment.id),
                  },
                  isDirectChild: true,
                  totals: running,

                  showCumulativeTimes,

                  setSegments,

                  onDelete,
                  onAddChild,
                  onUpdate,
              })
            : null;

        rows.push(
            <SegmentRow
                key={segment.id}
                segment={segment}
                depth={depth}
                index={i}
                inheritedGroup={inheritedGroup}
                isDirectChild={isDirectChild}
                average={displayAverage}
                pb={displayPB}
                gold={displayGold}
                onMoveUp={(id) => updateTree((prev) => moveSegmentUp(prev, id))}
                onMoveDown={(id) => updateTree((prev) => moveSegmentDown(prev, id))}
                onGroup={(id) => updateTree((prev) => groupIntoPreviousSibling(prev, id))}
                onUngroup={(id) => updateTree((prev) => ungroupToTopLevel(prev, id))}
                onDelete={onDelete}
                onAddChild={onAddChild}
                onUpdate={onUpdate}
            />,
        );

        if (!hasChildren) {
            running = {
                avg: addTime(running.avg, segment.average),
                pb: addTime(running.pb, segment.pb),
                gold: addTime(running.gold, segment.gold),
            };
        }

        if (childResult) {
            rows.push(...childResult.rows);
            running = childResult.totals;
        }
    }

    return {
        rows,
        totals: running,
    };
}
