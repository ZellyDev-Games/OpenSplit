import SegmentPayload from "../../../../models/segmentPayload";
import { TimeRow } from "../TimeRow";

type Props = {
    segment: SegmentPayload;

    average: number;
    pb: number;
    gold: number;
    showMilliseconds: boolean;

    onUpdate(id: string, updater: (segment: SegmentPayload) => SegmentPayload): void;
};

export default function SegmentTimeCells({ segment, average, pb, gold, showMilliseconds, onUpdate }: Props) {
    return (
        <>
            <td>
                <TimeRow
                    time={average}
                    showMilliseconds={showMilliseconds}
                    onChange={(value) =>
                        onUpdate(segment.id, (s) => ({
                            ...s,
                            average: value,
                        }))
                    }
                />
            </td>

            <td>
                <TimeRow
                    time={pb}
                    showMilliseconds={showMilliseconds}
                    onChange={(value) =>
                        onUpdate(segment.id, (s) => ({
                            ...s,
                            pb: value,
                        }))
                    }
                />
            </td>

            <td>
                <TimeRow
                    time={gold}
                    showMilliseconds={showMilliseconds}
                    onChange={(value) =>
                        onUpdate(segment.id, (s) => ({
                            ...s,
                            gold: value,
                        }))
                    }
                />
            </td>
        </>
    );
}
