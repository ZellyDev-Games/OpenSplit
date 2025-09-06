import {session} from "../../wailsjs/go/models";
import SplitFilePayload = session.SplitFilePayload;

type SplitListProps = {
    splitFile: SplitFilePayload | null
}

export default function SplitList({splitFile} : SplitListProps) {
    return(
    <div className="splitList">
        <ul>
            {splitFile && splitFile.segments.map((segment, index) =>
                <li key={segment.id ?? index}>
                    <div className="splitItem">
                        <p>
                            {segment.name}
                        </p>
                        <p>
                            {segment.average_time}
                        </p>
                    </div>
                </li>
            )}
        </ul>
    </div>
    )
}