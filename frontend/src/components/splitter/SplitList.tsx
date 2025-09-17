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
        if (index < completions.length) {
            let className = ""
            if(splitFile && completions[index].raw < splitFile.stats.golds[segment.id]?.raw)
            {
                className = "timer-gold"
            } else {
                if (splitFile) {
                    const target = compareAgainst == "average" ?
                        splitFile.stats.averages[segment.id].raw :
                        splitFile.stats.pb?.run?.split_payloads.find(s =>
                            s.split_segment_id === segment.id)?.current_time.raw
                    if (target) {
                        if (completions[index].raw > target) {
                            className = "timer-behind";
                        }

                        if (completions[index].raw < target) {
                            className = "timer-ahead";
                        }
                    }
                }
            }

            return (<strong className={className}>
                {completions[index].time}
            </strong>)
        } else {
            let val = ""
            let raw = 0
            if (compareAgainst == "average") {
                if (splitFile && splitFile.stats.averages[segment.id]) {
                    raw = splitFile.stats.averages[segment.id].raw;
                    val = splitFile.stats.averages[segment.id].formatted;
                } else {
                    return <strong>-</strong>
                }
            } else {
                const best = splitFile?.stats.pb?.run?.split_payloads.find((p) => p.split_segment_id === segment.id);
                if (best) {
                    raw = best.current_time.raw
                    val = best.current_time.formatted;
                } else {
                    return <strong>-</strong>;
                }
            }

            const diff = time - raw;
            let className = ""
            if(index === currentSegment && diff < 30000) {
                if (time < raw) {className = "timer-ahead"}
                if (time > raw) {className = "timer-behind"}
                return <strong className={className}>
                    {displayFormattedTimeParts(formatDuration(msToParts(diff), true))}
                </strong>
            }
            return <strong className={className}>{displayFormattedTimeParts(formatDuration(stringToParts(val)))}</strong>
        }
    };

    const segmentRows = splitFile?.segments.map((segment, index) => (
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
