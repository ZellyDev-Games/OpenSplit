import {session} from "../../wailsjs/go/models";
import SplitFilePayload = session.SplitFilePayload;
import {formatDuration, stringToParts} from "./Timer";

type SplitListProps = {
    splitFile: SplitFilePayload | null
    compareAgainst: CompareAgainst
}

export type CompareAgainst = "best" | "average"

export default function SplitList({splitFile, compareAgainst} : SplitListProps) {
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