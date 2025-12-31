import { JSX, useEffect, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";
import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";
import { displayFormattedTimeParts, formatDuration, msToParts } from "./Timer";

export type CompareAgainst = "best" | "average";

type Completion = {
    segmentID: string;
    time: string;
    raw: number;
};

type SplitListParameters = {
    sessionPayload: SessionPayload;
}

export default function SplitList({sessionPayload}: SplitListParameters) {
    const [completions, setCompletions] = useState<Completion[]>([]);
    const [compareAgainst] = useState<CompareAgainst>("average");
    const [time, setTime] = useState(0);

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
            const formatted = displayFormattedTimeParts(
                formatDuration(
                    msToParts(split.current_cumulative)));

            completed.push({
                segmentID: segmentID,
                time: `${formatted[0]}${formatted[1]}`,
                raw: split.current_duration
            })
        }
        setCompletions(completed);
    }, [sessionPayload]);

    // Segment time display
    const getSegmentDisplayTime = (leafIndex: number, segment: SegmentPayload): JSX.Element => {
        const gold = segment.gold;
        const average = segment.average;
        const best = segment.pb;

        const target = compareAgainst === "average" ? average : best;

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

        sessionPayload.loaded_split_file.segments.forEach((segment: SegmentPayload) => {
            // if this segment isn't in leaf segments it's a parent segment
            // in this case we don't show times
            const leafIndex = sessionPayload.leaf_segments?.indexOf(segment);
            if (leafIndex === -1 || leafIndex === undefined) {
                elements.push(<tr key={segment.id} className="parentRow">
                    <td className="splitName" style={{ /*paddingLeft: segment.depth * 16 */ }}>
                        <strong>{segment.name}</strong>
                    </td>
                    <td className="splitComparison"></td>
                </tr>)
            } else {
                const isSelected = leafIndex === sessionPayload.current_segment_index;
                elements.push(<tr key={segment.id} className={isSelected ? "selected" : ""}>
                    <td className="splitName" style={{ /*paddingLeft: seg.depth * 16 */}}>
                        {segment.name}
                    </td>

                    <td className="splitComparison">{getSegmentDisplayTime(leafIndex, segment)}</td>
                </tr>);
            }
        })

        return elements;
    }

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
