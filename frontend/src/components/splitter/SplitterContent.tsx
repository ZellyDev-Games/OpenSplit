import { useLayoutEffect, useState } from "react";

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
    hasSegmentIcons: boolean;
};

const comparisonLabel: Record<Comparison, string> = {
    [CompareAgainst.Average]: "Comparing Against: Average",
    [CompareAgainst.Best]: "Comparing Against: Best Run",
    [CompareAgainst.SumOfBest]: "Comparing Against: Sum of Best Segments",
};

type ColumnWidths = number[];

function SegmentColumnGroup({ widths, hasSegmentIcons }: { widths: number[] | null; hasSegmentIcons: boolean }) {
    return (
        <colgroup>
            {hasSegmentIcons && <col className="segmentIconCol" style={widths ? { width: widths[0] } : undefined} />}

            <col
                className="splitNameCol"
                style={
                    widths
                        ? {
                              width: widths[hasSegmentIcons ? 1 : 0],
                          }
                        : undefined
                }
            />

            <col
                className="splitDeltaCol"
                style={
                    widths
                        ? {
                              width: widths[hasSegmentIcons ? 2 : 1],
                          }
                        : undefined
                }
            />

            <col
                className="splitComparisonCol"
                style={
                    widths
                        ? {
                              width: widths[hasSegmentIcons ? 3 : 2],
                          }
                        : undefined
                }
            />

            <col
                className="splitTimeCol"
                style={
                    widths
                        ? {
                              width: widths[hasSegmentIcons ? 4 : 3],
                          }
                        : undefined
                }
            />
        </colgroup>
    );
}

export default function SplitterContent({
    sessionPayload,
    comparison,
    completeClassName,
    containerRef,
    rows,
    finalRow,
    hasSegmentIcons,
}: SplitterContentProps) {
    const [columnWidths, setColumnWidths] = useState<ColumnWidths | null>(null);

    useLayoutEffect(() => {
        const container = containerRef.current;

        if (!container) {
            return;
        }

        const table = container.querySelector<HTMLTableElement>("table");

        if (!table) {
            return;
        }

        const measureColumns = () => {
            const row = table.querySelector<HTMLTableRowElement>("tbody > tr");

            if (!row) {
                return;
            }

            const cells = Array.from(row.cells);
            const expectedCellCount = hasSegmentIcons ? 5 : 4;

            if (cells.length !== expectedCellCount) {
                return;
            }

            const widths = cells.map((cell) => cell.getBoundingClientRect().width);

            setColumnWidths((previous) => {
                if (
                    previous &&
                    previous.length === widths.length &&
                    previous.every((width, index) => Math.abs(width - widths[index]) < 0.5)
                ) {
                    return previous;
                }

                return widths;
            });
        };

        measureColumns();

        const row = table.querySelector<HTMLTableRowElement>("tbody > tr");

        if (!row) {
            return;
        }

        const resizeObserver = new ResizeObserver(measureColumns);

        resizeObserver.observe(table);

        for (const cell of row.cells) {
            resizeObserver.observe(cell);
        }

        return () => {
            resizeObserver.disconnect();
        };
    }, [containerRef, rows, hasSegmentIcons]);

    return (
        <>
            <SplitGameInfo sessionPayload={sessionPayload} completeClassName={completeClassName} />

            <div id="splitList" className={completeClassName}>
                <div ref={containerRef} id="splitContainer" className={completeClassName}>
                    <table cellSpacing={0} className={completeClassName}>
                        <SegmentColumnGroup widths={columnWidths} hasSegmentIcons={hasSegmentIcons} />
                        <tbody>{rows}</tbody>
                    </table>
                </div>

                <div id="finalSegment" className={completeClassName}>
                    <table cellSpacing={0} className={completeClassName}>
                        <SegmentColumnGroup widths={columnWidths} hasSegmentIcons={hasSegmentIcons} />
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
