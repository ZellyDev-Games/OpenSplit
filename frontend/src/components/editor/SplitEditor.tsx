import {
    faArrowDown,
    faArrowRightFromBracket,
    faArrowUp,
    faArrowUpFromBracket,
    faFolder,
    faTrash,
    IconDefinition,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import React, { useEffect, useRef, useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { WindowCenter, WindowSetSize } from "../../../wailsjs/runtime";
import { Command } from "../../App";
import { useClickOutside } from "../../hooks/useClickOutside";
import SegmentPayload from "../../models/segmentPayload";
import SplitFilePayload from "../../models/splitFilePayload";
import { FilePicker } from "../FilePicker";
import { msToParts, partsToMS, TimeParts } from "../splitter/Timer";
import TimeRow from "./TimeRow";

type GroupCtx = { bg: string };

function hashStringToInt(s: string): number {
    let h = 2166136261; // FNV-1a-ish
    for (let i = 0; i < s.length; i++) {
        h ^= s.charCodeAt(i);
        h = Math.imul(h, 16777619);
    }
    return h >>> 0;
}

function colorFromId(id: string): string {
    const n = hashStringToInt(id);

    const hue = n % 360;

    // Keep saturation strong but not neon
    const sat = 45 + (n % 15); // 45–59%

    // Dark background range
    const light = 18 + (n % 10); // 18–27%

    return `hsl(${hue} ${sat}% ${light}%)`;
}

type Game = {
    id: string;
    names: { international: string };
    assets: { "cover-tiny": { uri: string } };
    released: string;
};

type SplitEditorParams = {
    splitFilePayload: SplitFilePayload | null;
    speedRunAPIBase: string;
};

function addChildRecursive(list: SegmentPayload[], parent: SegmentPayload): SegmentPayload[] {
    return list.map((item) => {
        if (item.id === parent.id) {
            const child = new SegmentPayload();
            return {
                ...item,
                children: [...(item.children ?? []), child],
            };
        }

        return {
            ...item,
            children: addChildRecursive(item.children ?? [], parent),
        };
    });
}

// Clone to safely do in-place operations on the copy
function cloneSegments(list: SegmentPayload[]): SegmentPayload[] {
    return (list ?? []).map((seg) => {
        return new SegmentPayload({
            ...seg,
            children: cloneSegments(seg.children ?? []),
        });
    });
}

type ParentRef = { node: SegmentPayload; siblings: SegmentPayload[]; index: number };

function findNodeMutable(
    siblings: SegmentPayload[],
    id: string,
    parents: ParentRef[] = [],
): { siblings: SegmentPayload[]; index: number; parents: ParentRef[] } | null {
    for (let i = 0; i < siblings.length; i++) {
        const node = siblings[i];
        if (node.id === id) {
            return { siblings, index: i, parents };
        }
        const kids = node.children ?? [];
        if (kids.length > 0) {
            const nextParents = parents.concat([{ node, siblings, index: i }]);
            const found = findNodeMutable(kids, id, nextParents);
            if (found) return found;
        }
    }
    return null;
}

/**
 * Minimal tooltip wrapper:
 * - No external deps
 * - Accessible enough: uses title + aria-label + data-tooltip for CSS tooltip
 */
function IconButton({
    icon,
    onClick,
    tooltip,
    show = true,
}: {
    icon: IconDefinition;
    onClick: () => void;
    tooltip: string;
    show?: boolean;
}) {
    if (!show) return null;

    return (
        <button
            type="button"
            className="icon-btn has-tooltip"
            onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onClick();
            }}
            aria-label={tooltip}
            title={tooltip} // fallback if CSS isn't loaded
        >
            <FontAwesomeIcon icon={icon} />
            <span role="tooltip" className="tooltip-bubble">
                {tooltip}
            </span>
        </button>
    );
}

export default function SplitEditor({ splitFilePayload, speedRunAPIBase }: SplitEditorParams) {
    // Clear modal results which will close the modal when we click outside the modal
    const clickOutsideRef = useRef<HTMLDivElement | null>(null);
    useClickOutside(clickOutsideRef, () => {
        setGameResults([]);
    });

    // Segment stats
    const [splitFileLoaded] = useState<boolean>(false);
    const [gameName, setGameName] = React.useState<string>(splitFilePayload?.game_name ?? "");
    const [gameCategory, setGameCategory] = React.useState<string>(splitFilePayload?.game_category ?? "");
    const [attempts, setAttempts] = React.useState<number>(splitFilePayload?.attempts ?? 0);
    const [segments, setSegments] = useState<SegmentPayload[]>(splitFilePayload?.segments ?? []);
    const [offsetMS, setOffsetMS] = React.useState(0);
    const [autosplitterFile, setAutosplitterFile] = React.useState<string>(splitFilePayload?.autosplitter_file ?? "");

    // Speedrun search
    const [gameResults, setGameResults] = React.useState<Game[]>([]);
    const timeoutID = useRef<number>(0);

    // Position and size the edit window
    useEffect(() => {
        WindowSetSize(1000, 900);
        WindowCenter();
    }, []);

    useEffect(() => {
        (async () => {
            if (!splitFilePayload) return;
            console.log(splitFilePayload);
            setGameName(splitFilePayload.game_name);
            setGameCategory(splitFilePayload.game_category);
            setAttempts(splitFilePayload.attempts);
            setSegments(splitFilePayload.segments);
            setOffsetMS(splitFilePayload.offset);
        })();
    }, []);

    const searchSpeedrun = async () => {
        if (!speedRunAPIBase) return;
        const q = gameName.trim();
        if (!q) {
            setGameResults([]);
            return;
        }

        clearTimeout(timeoutID.current);
        const controller = new AbortController();
        timeoutID.current = setTimeout(async () => {
            fetch(`${speedRunAPIBase}/games?name=${encodeURIComponent(gameName)}`, {
                signal: controller.signal,
            })
                .then((res) => res.json())
                .then((data) => setGameResults(data.data))
                .catch((err) => {
                    if (err.name !== "AbortError") console.error("search failed:", err);
                });
        }, 500);
    };

    const addSegment = (parent: SegmentPayload | null) => {
        if (parent === null) {
            // top-level segment
            setSegments((prev) => [...prev, new SegmentPayload()]);
        } else {
            // subsegment
            setSegments((prev) => addChildRecursive(prev, parent));
        }
    };

    function updateSegmentName(id: string, name: string) {
        function updateRecursive(list: SegmentPayload[]): SegmentPayload[] {
            return list.map((item) => {
                if (item.id === id) {
                    return { ...item, name };
                }

                if ((item.children ?? []).length > 0) {
                    return {
                        ...item,
                        children: updateRecursive(item.children ?? []),
                    };
                }

                return item;
            });
        }

        setSegments((prev) => updateRecursive(prev));
    }

    const handleOffsetChange = (v: string) => {
        let t = parseInt(v, 10);
        if (isNaN(t)) {
            t = 0;
        }
        setOffsetMS(t);
    };

    const deleteSegment = (id: string) => {
        function deleteRecursive(list: SegmentPayload[]): SegmentPayload[] {
            return list
                .filter((seg) => seg.id !== id) // remove the target
                .map((seg) => ({
                    ...seg,
                    children: deleteRecursive(seg.children ?? []), // recurse downward
                }));
        }

        setSegments((prev) => deleteRecursive(prev));
    };

    const saveSplitFile = async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();

        const newSplitFilePayload = SplitFilePayload.createFrom({
            id: splitFilePayload?.id ?? "",
            version: splitFilePayload?.version ?? 0,
            runs: splitFilePayload?.runs ?? [],
            window_x: splitFilePayload?.window_x ?? 100,
            window_y: splitFilePayload?.window_y ?? 100,
            window_height: splitFilePayload?.window_height ?? 550,
            window_width: splitFilePayload?.window_width ?? 350,
            game_name: gameName,
            game_category: gameCategory,
            segments: segments,
            attempts: Number(attempts),
            pb: splitFilePayload?.pb ?? null,
            sob: splitFilePayload?.sob ?? 0,
            offset: offsetMS,
            autosplitter_file: autosplitterFile,
        });

        const payload = JSON.stringify(newSplitFilePayload);
        await Dispatch(Command.SUBMIT, payload);
    };

    const handleTimeChange = (id: string, time: TimeParts, isBest: boolean) => {
        const ms = partsToMS(time);

        function updateRecursive(list: SegmentPayload[]): SegmentPayload[] {
            return list.map((seg) => {
                if (seg.id === id) {
                    return new SegmentPayload({
                        ...seg,
                        pb: isBest ? ms : seg.pb,
                        average: isBest ? seg.average : ms,
                    });
                }

                if ((seg.children ?? []).length > 0) {
                    return new SegmentPayload({
                        ...seg,
                        children: updateRecursive(seg.children ?? []),
                    });
                }

                return seg;
            });
        }

        setSegments((prev) => updateRecursive(prev));
    };

    // 1) Move up/down among siblings
    const moveSegmentUp = (id: string) => {
        setSegments((prev) => {
            const root = cloneSegments(prev);
            const found = findNodeMutable(root, id);
            if (!found) return prev;

            const { siblings, index } = found;
            if (index <= 0) return prev;

            const tmp = siblings[index - 1];
            siblings[index - 1] = siblings[index];
            siblings[index] = tmp;

            return root;
        });
    };

    const moveSegmentDown = (id: string) => {
        setSegments((prev) => {
            const root = cloneSegments(prev);
            const found = findNodeMutable(root, id);
            if (!found) return prev;

            const { siblings, index } = found;
            if (index >= siblings.length - 1) return prev;

            const tmp = siblings[index + 1];
            siblings[index + 1] = siblings[index];
            siblings[index] = tmp;

            return root;
        });
    };

    // 2) Group into previous sibling
    const groupIntoPreviousSibling = (id: string) => {
        setSegments((prev) => {
            const root = cloneSegments(prev);
            const found = findNodeMutable(root, id);
            if (!found) return prev;

            const { siblings, index } = found;
            if (index <= 0) return prev;

            const node = siblings[index];
            const prevNode = siblings[index - 1];

            siblings.splice(index, 1);

            const prevChildren = prevNode.children ?? [];
            prevNode.children = [...prevChildren, node];

            return root;
        });
    };

    // 3) Ungroup to top-level
    const ungroupToTopLevel = (id: string) => {
        setSegments((prev) => {
            const root = cloneSegments(prev);
            const found = findNodeMutable(root, id);
            if (!found) return prev;

            const { siblings, index, parents } = found;
            if (parents.length === 0) return prev;

            const node = siblings[index];
            siblings.splice(index, 1);

            const topAncestor = parents[0].node;
            const topIndex = root.findIndex((s) => s.id === topAncestor.id);
            const insertAt = topIndex >= 0 ? topIndex + 1 : root.length;

            root.splice(insertAt, 0, node);

            return root;
        });
    };

    /**
     * renderRows arguments:
     * - depth: indent depth
     * - inheritedGroupShade: shading applied because this row is a direct child of a grouped parent
     */
    function renderRows(
        list: SegmentPayload[],
        depth: number,
        inheritedGroup: GroupCtx | null,
        isDirectChild: boolean,
    ) {
        return list.map((segment, i) => {
            const hasChildren = (segment.children ?? []).length > 0;

            // If THIS row is a group parent, it defines a new group color for itself + its direct children
            const ownGroup: GroupCtx | null = hasChildren ? { bg: colorFromId(segment.id) } : null;

            // Row styling:
            // - Group parent row: use its own group color
            // - Direct children of a group parent: use inherited group color
            // - Otherwise: no group styling
            const rowGroup: GroupCtx | null = ownGroup ?? (isDirectChild ? inheritedGroup : null);

            const rowStyle: React.CSSProperties | undefined = rowGroup
                ? ({ ["--group-bg"]: rowGroup.bg } as React.CSSProperties)
                : undefined;

            const inGroup = !!rowGroup;
            const isGroupParentRow = !!ownGroup;
            const isGroupChildRow = !ownGroup && isDirectChild && !!inheritedGroup;

            // Border box behavior:
            // - parent row: top border
            // - all rows in group: left/right
            // - last direct child row: bottom border
            const childIsLastInDirectGroup = isGroupChildRow && i === list.length - 1;

            const rowClassName = [
                inGroup ? "seg-group" : "",
                isGroupParentRow ? "seg-group-parent" : "",
                isGroupChildRow ? "seg-group-child" : "",
                childIsLastInDirectGroup ? "seg-group-bottom" : "",
            ]
                .filter(Boolean)
                .join(" ");

            // Only pass group context ONE level down:
            // - if segment is a group parent => its direct children inherit its group color
            // - otherwise => children do NOT inherit (unless they themselves become group parents)
            const nextInheritedGroup = ownGroup ?? null;
            const nextIsDirectChild = !!ownGroup;

            return (
                <React.Fragment key={segment.id}>
                    <tr className={rowClassName} style={rowStyle}>
                        <td>
                            <div style={{ display: "flex", flexDirection: "row", alignItems: "center", gap: 6 }}>
                                <IconButton
                                    icon={faArrowUp}
                                    tooltip="Move segment up"
                                    onClick={() => moveSegmentUp(segment.id)}
                                />
                                <IconButton
                                    icon={faArrowDown}
                                    tooltip="Move segment down"
                                    onClick={() => moveSegmentDown(segment.id)}
                                />
                                <IconButton
                                    icon={faArrowUpFromBracket}
                                    tooltip="Group under the segment above"
                                    show={i !== 0}
                                    onClick={() => groupIntoPreviousSibling(segment.id)}
                                />
                                <IconButton
                                    icon={faArrowRightFromBracket}
                                    tooltip="Remove from group (move to top level)"
                                    show={depth > 0}
                                    onClick={() => ungroupToTopLevel(segment.id)}
                                />
                            </div>
                        </td>

                        <td style={{ paddingLeft: depth * 20 }}>
                            <input
                                value={segment.name}
                                onChange={(e) => updateSegmentName(segment.id, e.target.value)}
                            />
                        </td>

                        <td>
                            {!hasChildren && (
                                <TimeRow
                                    id={segment.id}
                                    time={segment.average ? msToParts(segment.average) : null}
                                    onChangeCallback={(id, ts) => handleTimeChange(id, ts, false)}
                                />
                            )}
                        </td>

                        <td>
                            {!hasChildren && (
                                <TimeRow
                                    id={segment.id}
                                    time={segment.pb ? msToParts(segment.pb) : null}
                                    onChangeCallback={(id, ts) => handleTimeChange(id, ts, true)}
                                />
                            )}
                        </td>

                        <td>
                            <IconButton icon={faFolder} tooltip="Add subsegment" onClick={() => addSegment(segment)} />
                        </td>

                        <td>
                            <IconButton
                                icon={faTrash}
                                tooltip="Delete segment"
                                onClick={() => deleteSegment(segment.id)}
                            />
                        </td>
                    </tr>

                    {(segment.children ?? []).length > 0 &&
                        renderRows(segment.children ?? [], depth + 1, nextInheritedGroup, nextIsDirectChild)}
                </React.Fragment>
            );
        });
    }

    return (
        <div className="container form-container">
            <h2>{splitFileLoaded ? "Editing Split File" : "New Split File"}</h2>
            <form id="split-form" noValidate>
                <div className="row">
                    <label htmlFor="game_name">Game Name</label>
                    <input
                        value={gameName}
                        onChange={(e) => setGameName(e.target.value)}
                        onBlur={() => {
                            clearTimeout(timeoutID.current);
                        }}
                        onKeyUp={searchSpeedrun}
                        id="game_name"
                        name="game_name"
                        type="text"
                        autoComplete="off"
                    />
                </div>

                {gameResults.length > 0 && (
                    <div ref={clickOutsideRef} className="autocomplete">
                        <ul>
                            {gameResults.map((gameResult) => (
                                <li
                                    onClick={() => {
                                        setGameName(gameResult.names.international);
                                        setGameResults([]);
                                    }}
                                    key={gameResult.id}
                                >
                                    <div className="autocomplete-item">
                                        <img
                                            src={gameResult.assets["cover-tiny"].uri}
                                            alt={gameResult.assets["cover-tiny"].uri}
                                        />
                                        <div className="game-info">
                                            <strong>{gameResult.names.international}</strong>
                                            <span>{gameResult.released}</span>
                                        </div>
                                    </div>
                                </li>
                            ))}
                        </ul>
                    </div>
                )}

                <div className="row">
                    <label htmlFor="game_category">Category</label>
                    <input
                        onChange={(e) => setGameCategory(e.target.value)}
                        id="game_category"
                        name="game_category"
                        type="text"
                        autoComplete="off"
                        value={gameCategory}
                    />
                </div>

                <div className="row">
                    <label htmlFor="attempts">Attempts</label>
                    <input
                        onChange={(e) => setAttempts(Number(e.target.value))}
                        value={attempts ?? 0}
                        id="attempts"
                        name="attempts"
                        inputMode="numeric"
                    />
                </div>

                <div className="row">
                    <label htmlFor="offset">Negative Start Offset (milliseconds)</label>
                    <input
                        onChange={(e) => handleOffsetChange(e.target.value)}
                        id="offsetMS"
                        name="offsetMS"
                        type="text"
                        autoComplete="off"
                        value={offsetMS}
                    />
                </div>

                <div className="row">
                    <FilePicker fileName={autosplitterFile} setFilename={setAutosplitterFile} />
                </div>

                <div style={{ marginTop: 20, marginBottom: 20 }} className="row">
                    <div>
                        <button onClick={() => addSegment(null)} type="button">
                            Add Segment
                        </button>
                    </div>
                </div>

                <div className="datagrid-container">
                    <div className="datagrid">
                        {segments && segments.length > 0 && (
                            <table cellSpacing={0} className="datagrid" id="tbl-segments">
                                <thead>
                                    <tr>
                                        <th style={{ width: "5%" }}>#</th>
                                        <th style={{ width: "50%" }}>Segment Name</th>
                                        <th>
                                            Average Time <small>(HH:MM:SS.ccc)</small>
                                        </th>
                                        <th>
                                            Personal Best <small>(HH:MM:SS.ccc)</small>
                                        </th>
                                        <th style={{ width: "5%" }}>Add Subsegment</th>
                                        <th style={{ width: "5%" }}></th>
                                    </tr>
                                </thead>
                                <tbody>{renderRows(segments, 0, null, false)}</tbody>
                            </table>
                        )}
                    </div>
                </div>

                <hr />

                <div className="actions">
                    <button onClick={saveSplitFile} type="submit" className="primary">
                        Save
                    </button>
                    <button
                        type="button"
                        onClick={async () => {
                            await Dispatch(Command.CANCEL, null);
                        }}
                    >
                        Cancel
                    </button>
                </div>
            </form>
        </div>
    );
}
