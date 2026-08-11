import { JSX } from "react";

import SegmentPayload from "../../../models/segmentPayload";

type ParentSegmentRowProps = {
    segment: SegmentPayload;
    depth: number;
    completeClassName: string;
    isExpanded: boolean;
    hasChildren: boolean;
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
    const indentation = 5 + depth * 16;

    return (
        <tr className="parentRow">
            <td
                className="segmentIcon"
                style={{
                    paddingLeft: indentation,
                }}
            >
                {segment.icon && <img src={segment.icon} alt="" draggable={false} className="segment-icon" />}
            </td>

            <td
                className={"splitName parentName " + completeClassName}
                style={{
                    paddingLeft: indentation,
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

                <strong>{segment.name}</strong>
            </td>

            <td className={"splitDelta parentDelta " + completeClassName}>{parentDelta}</td>

            <td className={"splitComparison parentComparison " + completeClassName}>{parentSegmentDelta}</td>

            <td className={"splitTime parentTime " + completeClassName}>{parentComparison}</td>
        </tr>
    );
}
