import { session } from "../../../wailsjs/go/models";
import {displayFormattedTimeParts, formatDuration, stringToParts} from "./Timer";
import { useEffect, useState } from "react";
import { GetLoadedSplitFile, GetSessionStatus } from "../../../wailsjs/go/session/Service";
import { EventsOn } from "../../../wailsjs/runtime";
import SplitFilePayload = session.SplitFilePayload;
import ServicePayload = session.ServicePayload;

export type CompareAgainst = "best" | "average";

type Completion = {
    time: string;
}

export default function SplitList() {
    const [currentSegment, setCurrentSegment] = useState<number | null>(null);
    const [splitFile, setSplitFile] = useState<session.SplitFilePayload | null>(null);
    const [completions, setCompletions] = useState<Completion[]>([]);
    const [compareAgainst, setCompareAgainst] = useState<CompareAgainst | null>(null);

    useEffect(() => {
        (async () => {
            console.log("fetching loaded splitfile...");
            setSplitFile(await GetLoadedSplitFile());
        })();

        (async () => {
            console.log("fetching session data...");
            const session = await GetSessionStatus();
            if (session.current_segment !== undefined) {
                setCurrentSegment(session.current_segment_index);
            }
        })();

        const unsubscribeFromSplitUpdates = EventsOn("session:update", (servicePayload: ServicePayload) => {
            console.log("received service update:", servicePayload);
            setCurrentSegment(servicePayload.current_segment_index);
            if (servicePayload.current_run) {
                setCompletions(
                    servicePayload.current_run.split_payloads.map((c, i) => {
                        return {time: displayFormattedTimeParts(formatDuration(stringToParts(c.current_time)))}
                    }));
            } else {
                setCompletions([])
            }
        });

        const unsubscribeFromSplitFileUpdates = EventsOn("splitfile:update", (splitFilePayload: SplitFilePayload) => {
            console.log("received splitfile update", splitFilePayload);
            setSplitFile(splitFilePayload);
        });

        return () => {
            unsubscribeFromSplitFileUpdates();
            unsubscribeFromSplitUpdates();
        };
    }, []);

    const getSegmentDisplayTime = (index: number): string => {
        if(index < completions.length) {
            return completions[index].time;
        } else {
            console.log(splitFile?.segments)
            const avg = splitFile?.segments[index].average_time
            const best = splitFile?.segments[index].best_time
            return compareAgainst == "best" ? displayFormattedTimeParts(formatDuration(stringToParts(best ?? "-"))) :
                displayFormattedTimeParts(formatDuration(stringToParts(avg ?? "-")));
        }
    }

    const segmentRows = splitFile?.segments.map((segment, index) => (
        <tr key={segment.id ?? index} className={currentSegment !== null && currentSegment === index ? "selected" : ""}>
            <td className="splitName">{segment.name}</td>
            <td className="splitComparison">
                <strong>{getSegmentDisplayTime(index)}</strong>
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
