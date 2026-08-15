import { Dispatch, SetStateAction } from "react";

import SegmentPayload from "../../../models/segmentPayload";
import { SegmentUpdater } from "../types/segment";
import { renderSegmentRows } from "./SegmentRenderer";

type SegmentTableProps = {
    segments: SegmentPayload[];

    setSegments: Dispatch<SetStateAction<SegmentPayload[]>>;

    showCumulativeTimes: boolean;
    setShowCumulativeTimes: Dispatch<SetStateAction<boolean>>;

    onAddSegment: (parent: SegmentPayload | null) => void;
    onDeleteSegment: (id: string) => void;
    onUpdateSegment: SegmentUpdater;
};

export default function SegmentTable({
    segments,
    setSegments,
    showCumulativeTimes,
    setShowCumulativeTimes,
    onAddSegment,
    onDeleteSegment,
    onUpdateSegment,
}: SegmentTableProps) {
    return (
        <>
            <div
                style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    marginTop: 20,
                    marginBottom: 20,
                }}
            >
                <button type="button" onClick={() => onAddSegment(null)}>
                    Add Segment
                </button>

                <button type="button" onClick={() => setShowCumulativeTimes((value) => !value)}>
                    {showCumulativeTimes ? "Show Segment Times" : "Show Cumulative Times"}
                </button>
            </div>

            <div className="datagrid-container">
                <div className="datagrid-scroll">
                    {segments.length > 0 && (
                        <table id="tbl-segments" className="datagrid" cellSpacing={0}>
                            <thead>
                                <tr>
                                    <th style={{ width: "5%" }}>#</th>

                                    <th style={{ width: "12%" }}>Icon</th>

                                    <th style={{ width: "45%" }}>Segment Name</th>

                                    <th className="time-column">
                                        Average Time
                                        <small>(HH:MM:SS.cc)</small>
                                    </th>

                                    <th className="time-column">
                                        Personal Best
                                        <small>(HH:MM:SS.cc)</small>
                                    </th>

                                    <th className="time-column">
                                        Gold
                                        <small>(HH:MM:SS.cc)</small>
                                    </th>

                                    <th style={{ width: "5%" }}>Add</th>

                                    <th style={{ width: "5%" }} />
                                </tr>
                            </thead>

                            <tbody>
                                {
                                    renderSegmentRows({
                                        segments,
                                        setSegments,

                                        showCumulativeTimes,

                                        onDelete: onDeleteSegment,
                                        onAddChild: onAddSegment,
                                        onUpdate: onUpdateSegment,
                                    }).rows
                                }
                            </tbody>
                        </table>
                    )}
                </div>
            </div>
        </>
    );
}
