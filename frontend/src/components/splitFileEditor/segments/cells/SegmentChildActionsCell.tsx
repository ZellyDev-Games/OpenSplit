import { faFolder, faTrash } from "@fortawesome/free-solid-svg-icons";

import SegmentPayload from "../../../../models/segmentPayload";
import { IconButton } from "../../shared/IconButton";

type SegmentChildActionsCellProps = {
    segment: SegmentPayload;

    onAddChild(segment: SegmentPayload): void;

    onDelete(id: string): void;
};

export default function SegmentChildActionsCell({ segment, onAddChild, onDelete }: SegmentChildActionsCellProps) {
    return (
        <>
            <td>
                <IconButton icon={faFolder} tooltip="Add subsegment" onClick={() => onAddChild(segment)} />
            </td>

            <td>
                <IconButton icon={faTrash} tooltip="Delete segment" onClick={() => onDelete(segment.id)} />
            </td>
        </>
    );
}
