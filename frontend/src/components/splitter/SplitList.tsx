import { session } from "../../../wailsjs/go/models";
import { displayFormattedTimeParts, formatDuration, stringToParts } from "./Timer";
import { useEffect, useState } from "react";
import { GetLoadedSplitFile, GetSessionStatus } from "../../../wailsjs/go/session/Service";
import { EventsOn } from "../../../wailsjs/runtime";
import SplitFilePayload = session.SplitFilePayload;
import ServicePayload = session.ServicePayload;
import SegmentPayload = session.SegmentPayload;

export type CompareAgainst = "best" | "average";

type Completion = {
    time: string;
};

export default function SplitList() {
    const [splitFile, setSplitFile] = useState<SplitFilePayload | undefined>(undefined);
    const [currentSegment, setCurrentSegment] = useState<number | null>(null);
    const [completions, setCompletions] = useState<Completion[]>([]);
    const [compareAgainst, setCompareAgainst] = useState<CompareAgainst>("average");

    useEffect(() => {
        GetLoadedSplitFile().then((d) => setSplitFile(d));
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
                    servicePayload.current_run.split_payloads.map((c, i) => {
                        return { time: displayFormattedTimeParts(formatDuration(stringToParts(c.current_time))) };
                    }),
                );
            } else {
                setCompletions([]);
            }
        });
    }, []);

    const getSegmentDisplayTime = (index: number, segment: SegmentPayload): string => {
        if (index < completions.length) {
            return completions[index].time;
        } else {
            if (compareAgainst == "average") {
                const avg = splitFile?.Stats.averages[segment.id];
                if (avg) {
                    return displayFormattedTimeParts(formatDuration(stringToParts(avg))) ?? "-";
                } else {
                    return "-";
                }
            } else {
                const best = splitFile?.Stats.pb?.run?.split_payloads.find((p) => p.split_segment_id === segment.id);
                if (best) {
                    return displayFormattedTimeParts(formatDuration(stringToParts(best.current_time))) ?? "-";
                } else {
                    return "-";
                }
            }
        }
    };

    const segmentRows = splitFile?.segments.map((segment, index) => (
        <tr key={segment.id ?? index} className={currentSegment !== null && currentSegment === index ? "selected" : ""}>
            <td className="splitName">{segment.name}</td>
            <td className="splitComparison">
                <strong>{getSegmentDisplayTime(index, segment)}</strong>
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
