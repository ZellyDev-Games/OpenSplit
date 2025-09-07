import {session} from "../../../wailsjs/go/models";
import {formatDuration, stringToParts} from "./Timer";
import {useEffect, useState} from "react";
import {GetLoadedSplitFile} from "../../../wailsjs/go/session/Service";
import {EventsOn} from "../../../wailsjs/runtime";
import SplitFilePayload = session.SplitFilePayload;

export type CompareAgainst = "best" | "average"

export default function SplitList() {
    const [splitFile, setSplitFile] = useState<session.SplitFilePayload | null>(null);
    const [compareAgainst, setCompareAgainst] = useState<CompareAgainst | null>(null);

    useEffect(() => {
        (async() => {
            console.log("fetching loaded splitfile...")
            setSplitFile(await GetLoadedSplitFile())
        })()

        return EventsOn("splitfile:update", (splitFilePayload: SplitFilePayload) => {
            console.log("received splitfile update", splitFilePayload);
            setSplitFile(splitFilePayload)
        })
    }, [])

    const formattedSegments = splitFile?.segments.map((segment, index) =>{
        const time = compareAgainst === "average" ? segment.average_time : segment.best_time;
        const formattedParts = formatDuration(stringToParts(time))
        return `${formattedParts.hoursText}${formattedParts.sepHM}${formattedParts.minutesText}${formattedParts.sepMS}${formattedParts.secondsText}`;
    })

    const segmentRows = splitFile?.segments.map((segment, index) =>
        <tr key={segment.id ?? index}>
            <td className="splitName">
                {segment.name}
            </td>
            <td className="splitComparison">
                <strong>{formattedSegments && formattedSegments[index]}</strong>
            </td>
        </tr>
    )
    
    const rows = Array.isArray(segmentRows) ? segmentRows : [];
    const displayRows = rows.slice(0, -1);
    const finalRow = rows.at(-1) ?? null;

    return(
    <div className="splitList">
        <div className="gameInfo">
            <h1 className="gameTitle"><strong>{splitFile?.game_name}</strong></h1>
            <h2 className="gameCategory"><small>{splitFile?.game_category}</small></h2>
        </div>
        <div className="splitContainer">
            <table cellSpacing="0">
                <tbody>
                {displayRows}
                </tbody>
            </table>
        </div>
        <div className="finalSegment">
            <table><tbody>{finalRow}</tbody></table>
        </div>
    </div>
    )
}