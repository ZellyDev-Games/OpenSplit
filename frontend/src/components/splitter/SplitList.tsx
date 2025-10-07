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
};

export default function SplitList({ sessionPayload }: SplitListParameters) {
    const [completions, setCompletions] = useState<Completion[]>([]);
    const [compareAgainst] = useState<CompareAgainst>("average");
    const [time, setTime] = useState(0);

    useEffect(() => {
        return EventsOn("timer:update", (val: number) => {
            setTime(val);
        });
    }, []);

    useEffect(() => {
        if (sessionPayload?.current_run) {
            const completions: Completion[] = [];
            sessionPayload.current_run.splits.forEach((c) => {
                if (c != null) {
                    const time = displayFormattedTimeParts(formatDuration(msToParts(c.current_cumulative)));
                    completions.push({
                        segmentID: c.split_segment_id,
                        time: `${time[0]}${time[1]}`,
                        raw: c.current_duration,
                    });
                }
            });

            setCompletions(completions);
        } else {
            setCompletions([]);
        }
    }, [sessionPayload]);

    const getSegmentDisplayTime = (index: number, segment: SegmentPayload): JSX.Element => {
        const gold = segment.gold;
        const average = segment.average;
        const best = segment.pb;
        const target = compareAgainst == "average" ? average : best;

        const completion = completions.find((comp) => {
            return comp.segmentID === segment.id;
        });

        if (completion) {
            let className = "";
            if (gold && completion.raw < gold) {
                className = "timer-gold";
            } else {
                if (target) {
                    if (completion.raw > target) {
                        className = "timer-behind";
                    }

                    if (completion.raw < target) {
                        className = "timer-ahead";
                    }
                }
            }
            return <strong className={className}>{completion.time}</strong>;
        } else {
            if (target === 0) {
                return <>-</>;
            }

            const diff = time - target;
            let className = "";

            if (index === sessionPayload.current_segment_index && diff > -30000) {
                if (time < target) {
                    className = "timer-ahead";
                }
                if (time > target) {
                    className = "timer-behind";
                }
                const t = displayFormattedTimeParts(formatDuration(msToParts(diff), true));
                return (
                    <strong className={className}>
                        {`${t[0]}`}
                        <small>{`${t[1]}`}</small>
                    </strong>
                );
            }

            const t = displayFormattedTimeParts(formatDuration(msToParts(target)));
            return (
                <strong className={className}>
                    {`${t[0]}`}
                    <small>{`${t[1]}`}</small>
                </strong>
            );
        }
    };

    const segmentRows =
        sessionPayload.loaded_split_file &&
        sessionPayload.loaded_split_file.segments.map((segment, index) => (
            <tr key={segment.id ?? index} className={sessionPayload.current_segment_index === index ? "selected" : ""}>
                <td className="splitName">{segment.name}</td>
                <td className="splitComparison">{getSegmentDisplayTime(index, segment)}</td>
            </tr>
        ));

    const rows = Array.isArray(segmentRows) ? segmentRows : [];
    const displayRows = rows.slice(0, -1);
    const finalRow = rows.at(-1) ?? null;

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
                    <tbody>{displayRows}</tbody>
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
