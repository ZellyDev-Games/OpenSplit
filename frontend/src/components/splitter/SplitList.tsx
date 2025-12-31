import { JSX, useEffect, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";
import SegmentPayload from "../../models/segmentPayload";
import SessionPayload from "../../models/sessionPayload";
import { displayFormattedTimeParts, formatDuration, msToParts } from "./Timer";

export type CompareAgainst = "best" | "average";

type Completion = {
    segmentID: string;
    time: string;
    raw: number;
};

type SplitListParameters = {
    sessionPayload: SessionPayload;
};

type FlatSegment = SegmentPayload & {
    depth: number;
    isLeaf: boolean;
};

function flattenSegments(list: SegmentPayload[], depth = 0): FlatSegment[] {
    const out: FlatSegment[] = [];

    for (const seg of list) {
        const isLeaf = !seg.children || seg.children.length === 0;

        out.push({
            ...seg,
            depth,
            isLeaf,
        });

        if (!isLeaf) {
            out.push(...flattenSegments(seg.children, depth + 1));
        }
    }

    return out;
}

export default function SplitList({ sessionPayload }: SplitListParameters) {
    const [completions, setCompletions] = useState<Completion[]>([]);
    const [compareAgainst] = useState<CompareAgainst>("average");
    const [time, setTime] = useState(0);

    // subscribe to timer updates
    useEffect(() => {
        return EventsOn("timer:update", (val: number) => {
            setTime(val);
        });
    }, []);

    // completed splits from current run
    useEffect(() => {
        if (sessionPayload?.current_run) {
            const completed: Completion[] = [];

            sessionPayload.current_run.splits.forEach((c) => {
                if (c != null) {
                    const formatted = displayFormattedTimeParts(formatDuration(msToParts(c.current_cumulative)));

                    completed.push({
                        segmentID: c.split_segment_id,
                        time: `${formatted[0]}${formatted[1]}`,
                        raw: c.current_duration,
                    });
                }
            });

            setCompletions(completed);
        } else {
            setCompletions([]);
        }
    }, [sessionPayload]);

    // Segment time display
    const getSegmentDisplayTime = (leafIndex: number, segment: SegmentPayload): JSX.Element => {
        const gold = segment.gold;
        const average = segment.average;
        const best = segment.pb;

        const target = compareAgainst === "average" ? average : best;

        const completion = completions.find((comp) => comp.segmentID === segment.id);

        // Completed split
        if (completion) {
            let className = "";

            if (gold && completion.raw < gold) {
                className = "timer-gold";
            } else if (target) {
                if (completion.raw > target) className = "timer-behind";
                if (completion.raw < target) className = "timer-ahead";
            }

            return <strong className={className}>{completion.time}</strong>;
        }

        // Not completed yet
        if (!target) return <>-</>;

        const diff = time - target;

        // Show live ahead/behind only for current active leaf
        if (leafIndex === sessionPayload.current_segment_index && diff > -30000) {
            const t = displayFormattedTimeParts(formatDuration(msToParts(diff), true));
            const className = diff < 0 ? "timer-ahead" : "timer-behind";

            return (
                <strong className={className}>
                    {t[0]}
                    <small>{t[1]}</small>
                </strong>
            );
        }

        // Default target display
        const t = displayFormattedTimeParts(formatDuration(msToParts(target)));
        return (
            <strong>
                {t[0]}
                <small>{t[1]}</small>
            </strong>
        );
    };

    // row renderer
    const rows = sessionPayload.loaded_split_file?.segments.map((seg) => {
        if (sessionPayload.runtime_segments?.indexOf(seg) === -1) {
            return (
                <tr key={seg.id} className="parentRow">
                    <td className="splitName" style={{ paddingLeft: seg.depth * 16 }}>
                        <strong>{seg.name}</strong>
                    </td>
                    <td className="splitComparison"></td>
                </tr>
            );
        }

        // leaf/real segment
        const leafIndex = sessionPayload.runtime_segments?.findIndex((l) => l.id === seg.id);
        const isSelected = leafIndex === sessionPayload.current_segment_index;

        return (
            <tr key={seg.id} className={isSelected ? "selected" : ""}>
                <td className="splitName" style={{ paddingLeft: seg.depth * 16 }}>
                    {seg.name}
                </td>

                <td className="splitComparison">{getSegmentDisplayTime(leafIndex, seg)}</td>
            </tr>
        );
    });

    // Final row separated
    const leafRows = rows.filter((r) => r.props.className !== "parentRow");
    const finalRow = leafRows.at(-1);

    return (
        <div className="splitList">
            <div className="gameInfo">
                <h1 className="gameTitle">
                    <strong>{sessionPayload.loaded_split_file?.game_name}</strong>
                </h1>
                <h2 className="gameCategory">
                    <small>{sessionPayload.loaded_split_file?.game_category}</small>
                </h2>
            </div>

            <div className="splitContainer">
                <table cellSpacing="0">
                    <tbody>{rows}</tbody>
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
