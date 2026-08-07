import { CSSProperties } from "react";

import SegmentPayload from "../../../../models/segmentPayload";

type SegmentNameCellProps = {
    segment: SegmentPayload;

    depth: number;

    onUpdate(id: string, updater: (segment: SegmentPayload) => SegmentPayload): void;
};

export default function SegmentNameCell({ segment, depth, onUpdate }: SegmentNameCellProps) {
    return (
        <td
            className="segment-name-cell"
            style={
                {
                    "--segment-depth": depth,
                } as CSSProperties
            }
        >
            <input
                value={segment.name}
                onChange={(event) =>
                    onUpdate(segment.id, (current) => ({
                        ...current,
                        name: event.target.value,
                    }))
                }
            />
        </td>
    );
}
