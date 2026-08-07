import SegmentPayload from "../../../../models/segmentPayload";
import SegmentActions from "../SegmentActions";

type SegmentActionCellProps = {
    segment: SegmentPayload;

    index: number;

    depth: number;

    onMoveUp(id: string): void;

    onMoveDown(id: string): void;

    onGroup(id: string): void;

    onUngroup(id: string): void;
};

export default function SegmentActionCell({
    segment,
    index,
    depth,
    onMoveUp,
    onMoveDown,
    onGroup,
    onUngroup,
}: SegmentActionCellProps) {
    return (
        <td>
            <SegmentActions
                segment={segment}
                index={index}
                depth={depth}
                onMoveUp={onMoveUp}
                onMoveDown={onMoveDown}
                onGroup={onGroup}
                onUngroup={onUngroup}
            />
        </td>
    );
}
