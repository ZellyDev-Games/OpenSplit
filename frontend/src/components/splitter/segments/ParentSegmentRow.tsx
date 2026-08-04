import { JSX } from "react";

import SegmentPayload from "../../../models/segmentPayload";
import SplitPayload from "../../../models/splitPayload";

type ParentSegmentRowProps = {
    segment: SegmentPayload;
    depth: number;
    completeClassName: string;
    isExpanded: boolean;
    hasChildren: boolean;
    lastLeafSplit: SplitPayload | null;
    parentComparison: JSX.Element | null;
    parentDelta: JSX.Element | null;
    parentSegmentDelta: JSX.Element | null;
    onToggle: () => void;
};

export default function ParentSegmentRow({
    segment,
    depth,
    completeClassName,
    isExpanded,
    hasChildren,
    parentComparison,
    parentDelta,
    parentSegmentDelta,
    onToggle,
}: ParentSegmentRowProps) {
    return (
        <tr className="parentRow">
            <td
                className={"splitName " + completeClassName}
                style={{
                    paddingLeft: depth * 16,
                }}
            >
                {hasChildren && (
                    <button
                        type="button"
                        className="collapseToggle"
                        onClick={(event) => {
                            event.stopPropagation();
                            onToggle();
                        }}
                        aria-label={isExpanded ? "Collapse segment group" : "Expand segment group"}
                    >
                        {isExpanded ? "▾" : "▸"}
                    </button>
                )}

                {segment.icon && <img src={segment.icon} alt="" draggable={false} className="segment-icon" />}

                <strong>{segment.name}</strong>
            </td>

            <td className={"splitDelta " + completeClassName}>{parentDelta}</td>

            <td className={"splitComparison " + completeClassName}>{parentSegmentDelta}</td>

            <td className={"splitTime " + completeClassName}>{parentComparison}</td>
        </tr>
    );
}
