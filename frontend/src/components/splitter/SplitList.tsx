import { JSX, useEffect, useMemo, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";
import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";
import { CompareAgainst, Comparison } from "./Splitter";
import { displayFormattedTimeParts, formatDuration, msToParts } from "./Timer";

type Completion = {
    segmentID: string;
    time: string;
    raw: number;
};

type SplitListParameters = {
    sessionPayload: SessionPayload;
    comparison: Comparison;
};

type FlatSegment = {
    Segment: SegmentPayload;
    Depth: number;
};

function flattenSegments(segments: SegmentPayload[], depth: number = 0): FlatSegment[] {
    const flatSegments: FlatSegment[] = [];
    for (const segment of segments) {
        console.log(segment);
        flatSegments.push({
            Segment: segment,
            Depth: depth,
        });

        if (segment.children.length > 0) {
            flatSegments.push(...flattenSegments(segment.children, depth + 1));
        }
    }

    return flatSegments;
}

export default function SplitList({ sessionPayload, comparison }: SplitListParameters) {
    const [completions, setCompletions] = useState<Completion[]>([]);
    const [time, setTime] = useState(0);

    const flatSegments = useMemo<FlatSegment[]>(() => {
        if (!sessionPayload.loaded_split_file) {
            return [];
        }

        return flattenSegments(sessionPayload.loaded_split_file.segments);
    }, [sessionPayload.loaded_split_file]);

    // subscribe to timer updates
    useEffect(() => {
        return EventsOn("timer:update", (val: number) => {
            setTime(val);
        });
    }, []);

    // completed splits from current run
    useEffect(() => {
        if (!sessionPayload?.current_run) {
            setCompletions([]);
            return;
        }

        const completed: Completion[] = [];
        for (const segmentID of Object.keys(sessionPayload.current_run.splits)) {
            const split = sessionPayload.current_run.splits[segmentID];
            const formatted = displayFormattedTimeParts(formatDuration(msToParts(split.current_cumulative)));

            completed.push({
                segmentID: segmentID,
                time: `${formatted[0]}${formatted[1]}`,
                raw: split.current_duration,
            });
        }
        setCompletions(completed);
    }, [sessionPayload]);

    // Segment time display
    const getSegmentDisplayTime = (leafIndex: number, segment: SegmentPayload): JSX.Element => {
        const gold = segment.gold;
        const average = segment.average;
        const best = segment.pb;

        const target = comparison === CompareAgainst.Average ? average : best;

        const completion = completions.find((comp) => comp.segmentID === segment.id);

        // Completed split
        if (completion) {
            let className = "";

            if (gold && completion.raw < gold) {
                className = "timer-gold";
            } else if (target) {
                if (completion.raw > target) className = "timer-behind";
                if (completion.raw < target) className = "timer-ahead";
            }

            return <strong className={className}>{completion.time}</strong>;
        }

        // Not completed yet
        if (!target) return <>-</>;

        const diff = time - target;

        // Show live ahead/behind only for current active leaf
        if (leafIndex === sessionPayload.current_segment_index && diff > -30000) {
            const t = displayFormattedTimeParts(formatDuration(msToParts(diff), true));
            const className = diff < 0 ? "timer-ahead" : "timer-behind";

            return (
                <strong className={className}>
                    {t[0]}
                    <small>{t[1]}</small>
                </strong>
            );
        }

        // Default target display
        const t = displayFormattedTimeParts(formatDuration(msToParts(target)));
        return (
            <strong>
                {t[0]}
                <small>{t[1]}</small>
            </strong>
        );
    };

    // row renderer
    const rows = (): JSX.Element[] => {
        const elements: JSX.Element[] = [];
        if (!sessionPayload.loaded_split_file || sessionPayload.leaf_segments == null) {
            return [];
        }

        flatSegments.forEach((segmentData: FlatSegment) => {
            // if this segmentData isn't in leaf segments it's a parent segmentData
            // in this case we don't show times
            const leafIndex = sessionPayload.leaf_segments?.findIndex((leaf) => leaf.id == segmentData.Segment.id);
            if (leafIndex === -1 || leafIndex === undefined) {
                elements.push(
                    <tr key={segmentData.Segment.id} className="parentRow">
                        <td className="splitName" style={{ paddingLeft: segmentData.Depth * 16 }}>
                            <strong>{segmentData.Segment.name}</strong>
                        </td>
                        <td className="splitComparison"></td>
                    </tr>,
                );
            } else {
                const isSelected = leafIndex === sessionPayload.current_segment_index;
                console.log(leafIndex, sessionPayload.current_segment_index);
                elements.push(
                    <tr key={segmentData.Segment.id} className={isSelected ? "selected" : ""}>
                        <td className="splitName" style={{ paddingLeft: segmentData.Depth * 16 }}>
                            {segmentData.Segment.name}
                        </td>

                        <td className="splitComparison">{getSegmentDisplayTime(leafIndex, segmentData.Segment)}</td>
                    </tr>,
                );
            }
        });

        return elements;
    };

    // Final row separated
    const leafRows = rows().filter((r) => r.props.className !== "parentRow");
    const finalRow = leafRows.at(-1);

    return (
        <div className="splitList">
            <div className="gameInfo">
                <h1 className="gameTitle">
                    <strong>{sessionPayload.loaded_split_file?.game_name}</strong>
                </h1>
                <h2 className="gameCategory">
                    <small>{sessionPayload.loaded_split_file?.game_category}</small>
                </h2>
            </div>

            <div className="splitContainer">
                <table cellSpacing="0">
                    <tbody>{rows()}</tbody>
                </table>
            </div>

            <div className="finalSegment">
                <table>
                    <tbody>{finalRow}</tbody>
                </table>
            </div>
        </div>
    );
}
