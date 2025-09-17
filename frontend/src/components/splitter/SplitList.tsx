import { session } from "../../../wailsjs/go/models";
import {displayFormattedTimeParts, formatDuration, msToParts, stringToParts} from "./Timer";
import {JSX, useEffect, useState} from "react";
import { GetLoadedSplitFile, GetSessionStatus } from "../../../wailsjs/go/session/Service";
import { EventsOn } from "../../../wailsjs/runtime";
import SplitFilePayload = session.SplitFilePayload;
import ServicePayload = session.ServicePayload;
import SegmentPayload = session.SegmentPayload;

export type CompareAgainst = "best" | "average";

type Completion = {
    time: string;
    raw: number;
};

export default function SplitList() {
    const [splitFile, setSplitFile] = useState<SplitFilePayload | undefined>(undefined);
    const [currentSegment, setCurrentSegment] = useState<number | null>(null);
    const [completions, setCompletions] = useState<Completion[]>([]);
    const [compareAgainst, _] = useState<CompareAgainst>("average");
    const [time, setTime] = useState(0);

    useEffect(() => {
        GetLoadedSplitFile().then((d) => setSplitFile(d));
    }, []);

    useEffect(() => {
        return EventsOn("timer:update", (val: number) => {
            setTime(val);
        });
    }, []);

    useEffect(() => {
        (async () => {
            console.log("fetching session data...");
            const session = await GetSessionStatus();
            if (session.current_segment !== undefined) {
                setCurrentSegment(session.current_segment_index);
            }
        })();

        return EventsOn("session:update", (servicePayload: ServicePayload) => {
            console.log("received service update:", servicePayload);
            setSplitFile(servicePayload.split_file);
            setCurrentSegment(servicePayload.current_segment_index);
            if (servicePayload.current_run) {
                setCompletions(
                    servicePayload.current_run.split_payloads.map((c, _) => {
                        return {
                            time: displayFormattedTimeParts(formatDuration(stringToParts(c.current_time.formatted))),
                            raw: c.current_time.raw
                        };
                    }),
                );
            } else {
                setCompletions([]);
            }
        });
    }, []);

    const getSegmentDisplayTime = (index: number, segment: SegmentPayload): JSX.Element => {
        const gold = segment.gold?.raw
        const average = segment.average?.raw
        const best  = segment.pb?.raw
        const target = compareAgainst == "average" ? average : best

        if (index < completions.length) {
            let className = ""
            if(gold && completions[index].raw < gold)
            {
                className = "timer-gold"
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

            return (<strong className={className}>
                {completions[index].time}
            </strong>)
        } else {
            const diff = target - time;
            console.log(diff)
            let className = ""

            if(index === currentSegment && diff < 30000) {
                if (time < target) {className = "timer-ahead"}
                if (time > target) {className = "timer-behind"}
                return <strong className={className}>
                    {displayFormattedTimeParts(formatDuration(msToParts(diff), true))}
                </strong>
            }

            return <strong className={className}>{displayFormattedTimeParts(formatDuration(msToParts(target)))}</strong>
        }
    };

    const segmentRows = splitFile?.segments?.map((segment, index) => (
        <tr key={segment.id ?? index} className={currentSegment !== null && currentSegment === index ? "selected" : ""}>
            <td className="splitName">{segment.name}</td>
            <td className="splitComparison">
                {getSegmentDisplayTime(index, segment)}
            </td>
        </tr>
    ));

    const rows = Array.isArray(segmentRows) ? segmentRows : [];
    const displayRows = rows.slice(0, -1);
    const finalRow = rows.at(-1) ?? null;

    return (
        <div className="splitList">
            <div className="gameInfo">
                <h1 className="gameTitle">
                    <strong>{splitFile?.game_name}</strong>
                </h1>
                <h2 className="gameCategory">
                    <small>{splitFile?.game_category}</small>
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
