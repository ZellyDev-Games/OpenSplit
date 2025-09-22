import { JSX, useEffect, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";
import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";
import { displayFormattedTimeParts, formatDuration, msToParts, stringToParts } from "./Timer";

export type CompareAgainst = "best" | "average";

type Completion = {
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
            setCompletions(
                sessionPayload.current_run.split_payloads.map((c) => {
                    const time = displayFormattedTimeParts(formatDuration(stringToParts(c.current_time.formatted)));
                    return {
                        time: `${time[0]}${time[1]}`,
                        raw: c.current_time.raw,
                    };
                }),
            );
        } else {
            setCompletions([]);
        }
    }, [sessionPayload]);

    const getSegmentDisplayTime = (index: number, segment: SegmentPayload): JSX.Element => {
        const gold = segment.gold?.raw;
        const average = segment.average?.raw;
        const best = segment.pb?.raw;
        const target = compareAgainst == "average" ? average : best;

        if (index < completions.length) {
            let className = "";
            if (gold && completions[index].raw < gold) {
                className = "timer-gold";
            } else {
                if (target) {
                    if (completions[index].raw > target) {
                        className = "timer-behind";
                    }

                    if (completions[index].raw < target) {
                        className = "timer-ahead";
                    }
                }
            }

            return <strong className={className}>{completions[index].time}</strong>;
        } else {
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
        sessionPayload.split_file &&
        sessionPayload.split_file.segments.map((segment, index) => (
            <tr
                key={segment.id ?? index}
                className={
                    sessionPayload.current_segment !== null && sessionPayload.current_segment_index === index
                        ? "selected"
                        : ""
                }
            >
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
                    <strong>{sessionPayload.split_file?.game_name}</strong>
                </h1>
                <h2 className="gameCategory">
                    <small>{sessionPayload.split_file?.game_category}</small>
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
