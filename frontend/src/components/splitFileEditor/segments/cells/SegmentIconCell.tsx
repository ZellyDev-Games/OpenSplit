import SegmentPayload from "../../../../models/segmentPayload";
import SegmentIconPicker from "../SegmentIconPicker";

type SegmentIconCellProps = {
    segment: SegmentPayload;

    onUpdate(id: string, updater: (segment: SegmentPayload) => SegmentPayload): void;
};

export default function SegmentIconCell({ segment, onUpdate }: SegmentIconCellProps) {
    return (
        <td>
            <SegmentIconPicker
                icon={segment.icon}
                onChange={(icon) =>
                    onUpdate(segment.id, (current) => ({
                        ...current,
                        icon,
                    }))
                }
            />
        </td>
    );
}
