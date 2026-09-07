import { CompareAgainst, Comparison } from "../../hooks/splitter/useComparison";
import SessionPayload from "../../models/sessionPayload";
import SplitGameInfo from "./SplitGameInfo";
import Timer from "./timer/Timer";
import WorldRecordDisplay from "./timer/WorldRecord";

type SplitterContentProps = {
    sessionPayload: SessionPayload;
    comparison: Comparison;
    forceExpandAll: boolean;
    completeClassName: string;
    containerRef: React.RefObject<HTMLDivElement | null>;
    rows: React.ReactNode;
    finalRow: React.ReactNode;
};

const comparisonLabel: Record<Comparison, string> = {
    [CompareAgainst.Average]: "Comparing Against: Average",
    [CompareAgainst.Best]: "Comparing Against: Best Run",
    [CompareAgainst.SumOfBest]: "Comparing Against: Sum of Best Segments",
};

export default function SplitterContent({
    sessionPayload,
    comparison,
    completeClassName,
    containerRef,
    rows,
    finalRow,
}: SplitterContentProps) {
    return (
        <>
            <SplitGameInfo sessionPayload={sessionPayload} completeClassName={completeClassName} />

            <div id="splitList" className={completeClassName}>
                <div ref={containerRef} id="splitContainer" className={completeClassName}>
                    <table cellSpacing={0} className={completeClassName}>
                        <tbody>{rows}</tbody>
                    </table>
                </div>

                <div id="finalSegment" className={completeClassName}>
                    <table cellSpacing={0} className={completeClassName}>
                        <tbody>{finalRow}</tbody>
                    </table>
                </div>
            </div>

            <div id="splitterInfo">
                <div className="comparison-mode">{comparisonLabel[comparison]}</div>

                <Timer offset={sessionPayload.loaded_split_file?.offset ?? 0} />

                <WorldRecordDisplay worldRecord={sessionPayload.loaded_split_file?.wr} />
            </div>
        </>
    );
}
