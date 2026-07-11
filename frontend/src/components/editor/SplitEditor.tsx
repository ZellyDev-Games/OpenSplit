import {
    faArrowDown,
    faArrowRightFromBracket,
    faArrowUp,
    faArrowUpFromBracket,
    faFolder,
    faTrash,
} from "@fortawesome/free-solid-svg-icons";
import React, { useEffect, useState } from "react";

import { Dispatch, ExportSplitFile } from "../../../wailsjs/go/dispatcher/Service";
import { GetAvailableSkins } from "../../../wailsjs/go/skin/Service";
import addIcon from "../../assets/images/add.png";
import removeIcon from "../../assets/images/remove.png";
import { Platforms, SearchCategories, SearchGames } from "../../../wailsjs/go/speedrun/Service";
import { WindowCenter, WindowSetSize } from "../../../wailsjs/runtime";
import { Command } from "../../App";
import { useClickOutside } from "../../hooks/useClickOutside";
import SegmentPayload from "../../models/segmentPayload";
import SplitFilePayload from "../../models/splitFilePayload";
import { IconButton } from "../Tooltip";
import { colorFromId, GroupCtx } from "./hashColor";
import { TimeRow } from "./TimeRow";

type SplitEditorParams = {
    splitFilePayload: SplitFilePayload | null;
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

type GameMatch = {
    id: string;
    name: string;
    platforms: Platform[];
};

type Platform = {
    id: string;
    name: string;
};

type Category = {
    id: string;
    name: string;
};

export default function SplitEditor({ splitFilePayload }: SplitEditorParams) {
    // Is this a new file or are we editing?
    const editing = splitFilePayload != null;

    const [platform, setPlatform] = React.useState<string>(splitFilePayload?.platform ?? "SNES");
    const [platforms, setPlatforms] = useState<Platform[]>([]);

    const [gameName, setGameName] = React.useState<string>(splitFilePayload?.game_name ?? "");
    const [gameID, setGameID] = useState(splitFilePayload?.speedrun_game_id ?? "");
    const [games, setGames] = useState<GameMatch[]>([]);

    const [categoryName, setCategoryName] = React.useState<string>(splitFilePayload?.game_category ?? "");
    const [CategoryID, setCategoryID] = useState(splitFilePayload?.speedrun_game_category_id ?? "");
    const [categories, setCategories] = useState<Category[]>([]);

    const [attempts, setAttempts] = React.useState<number>(splitFilePayload?.attempts ?? 0);
    const [segments, setSegments] = useState<SegmentPayload[]>(splitFilePayload?.segments ?? []);
    const [offsetMS, setOffsetMS] = useState(splitFilePayload?.offset ?? 0);
    const [offsetText, setOffsetText] = useState(String(splitFilePayload?.offset ?? 0));
    const [availableSkins, setAvailableSkins] = useState<string[]>([]);
    const [selectedSkin, setSelectedSkin] = useState(splitFilePayload?.selected_skin ?? "");

    const [showCumulativeTimes, setShowCumulativeTimes] = useState(false);
    const [gameActive, setGameActive] = useState(false);
    const [categoryActive, setCategoryActive] = useState(false);
    const selectingGame = React.useRef(false);

    useEffect(() => {
        if (selectingGame.current) {
            selectingGame.current = false;
            return;
        }

        const query = gameName.trim();

        const timeout = setTimeout(async () => {
            if (query.length === 0) {
                setGames([]);
                // setShowGames(false);
                return;
            }

            const games = await SearchGames(query);

            setGames(
                games.data.map((g) => ({
                    id: g.id,
                    name: g.names.international,
                    platforms: g.platforms,
                })),
            );

            // setShowGames(gameInputFocused && query.length > 0);
        }, 200);

        return () => clearTimeout(timeout);
    }, [gameName]);

    useEffect(() => {
        console.log("gameID changed:", gameID);

        if (!gameID) {
            setCategories([]);
            // setShowCategories(false);
            return;
        }

        SearchCategories(gameID).then((result) => {
            console.log("categories", result);

            setCategories(result.data);
            // setShowCategories(gameInputFocused && result.data.length > 0);
        });
    }, [gameID]);

    useEffect(() => {
        Platforms().then((p) => {
            setPlatforms(p);
        });
    }, []);

    useEffect(() => {
        const offset = splitFilePayload?.offset ?? 0;
        setOffsetMS(offset);
        setOffsetText(String(offset));
    }, [splitFilePayload]);

    const handleOffsetChange = (v: string) => {
        // Allow empty string, "-" and any integer being typed
        if (/^-?\d*$/.test(v)) {
            setOffsetText(v);
        }
    };

    useEffect(() => {
        GetAvailableSkins().then((skins) => {
            setAvailableSkins(skins);

            if (splitFilePayload?.selected_skin) {
                setSelectedSkin(splitFilePayload.selected_skin);
            } else if (skins.length > 0) {
                setSelectedSkin(skins[0]);
            }
        });
    }, [splitFilePayload]);

    useEffect(() => {
        console.log("SplitEditor payload", JSON.stringify(splitFilePayload, null, 2));
        console.log("offsetMS state", offsetMS);
    }, [splitFilePayload]);

    // Position and size the edit window
    useEffect(() => {
        WindowSetSize(1000, 900);
        WindowCenter();
    }, []);

    const addSegment = (parent: SegmentPayload | null) => {
        if (parent === null) {
            // top-level segment
            setSegments((prev) => [...prev, new SegmentPayload()]);
        } else {
            // subsegment
            setSegments((prev) => addChildRecursive(prev, parent));
        }
    };

    const updateSegmentIcon = (id: string, event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file) {
            return;
        }

        const reader = new FileReader();

        reader.onload = () => {
            updateSegment(id, (segment) => ({
                ...segment,
                icon: reader.result as string,
            }));
        };

        reader.readAsDataURL(file);

        // allow selecting the same file again later
        event.target.value = "";
    };

    function updateSegment(id: string, updater: (segment: SegmentPayload) => SegmentPayload) {
        function updateRecursive(list: SegmentPayload[]): SegmentPayload[] {
            return list.map((item) => {
                if (item.id === id) {
                    return updater(item);
                }

                return {
                    ...item,
                    children: updateRecursive(item.children ?? []),
                };
            });
        }

        setSegments((prev) => updateRecursive(prev));
    }

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

    const emptyWR = {
        show: false,
        run_id: "",
        players: [],
        real_time: 0,
        in_game_time: 0,
    };

    const saveSplitFile = async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();

        const newSplitFilePayload = SplitFilePayload.createFrom({
            id: splitFilePayload?.id ?? "",
            game_name: gameName,
            speedrun_game_id: gameID,
            game_category: categoryName,
            speedrun_game_category_id: CategoryID,
            version: splitFilePayload?.version ?? 0,

            selected_skin: selectedSkin,

            segments: segments,
            runs: splitFilePayload?.runs ?? [],
            pb: splitFilePayload?.pb ?? null,

            sob: splitFilePayload?.sob ?? 0,
            attempts: Number(attempts),
            offset: offsetText === "" || offsetText === "-" ? 0 : parseInt(offsetText, 10),
            platform: platform,

            wr: splitFilePayload?.wr ?? emptyWR,

            window_x: splitFilePayload?.window_x ?? 100,
            window_y: splitFilePayload?.window_y ?? 100,
            window_height: splitFilePayload?.window_height ?? 550,
            window_width: splitFilePayload?.window_width ?? 350,
        });

        const payload = JSON.stringify(newSplitFilePayload);
        await Dispatch(Command.SUBMIT, payload);
    };

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
  
    function updateSegmentTimes(id: string, average: number, pb: number) {
        function update(list: SegmentPayload[]): SegmentPayload[] {
            return list.map((seg) => {
                if (seg.id === id) {
                    return {
                        ...seg,
                        average,
                        pb,
                    };
                }

                return {
                    ...seg,
                    children: update(seg.children ?? []),
                };
            });
        }

        setSegments((prev) => update(prev));
    }

    const gameAutocompleteRef = React.useRef<HTMLDivElement>(null);
    const categoryAutocompleteRef = React.useRef<HTMLDivElement>(null);

    useClickOutside(gameAutocompleteRef, () => setGameActive(false));
    useClickOutside(categoryAutocompleteRef, () => setCategoryActive(false));

    const longestGameWidth = React.useMemo(() => {
        const canvas = document.createElement("canvas");
        const ctx = canvas.getContext("2d");
        if (!ctx) return 250;

        ctx.font = getComputedStyle(document.body).font;

        return Math.max(150, ...games.map((m) => ctx.measureText(m.name).width + 10));
    }, [games]);

    const longestCategoryWidth = React.useMemo(() => {
        const canvas = document.createElement("canvas");
        const ctx = canvas.getContext("2d");
        if (!ctx) return 150;

        ctx.font = getComputedStyle(document.body).font;

        return Math.max(150, ...categories.map((c) => ctx.measureText(c.name).width + 10));
    }, [categories]);

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
        totalAvg: number,
        totalPB: number,
    ): {
        rows: React.ReactElement[];
        totalAvg: number;
        totalPB: number;
    } {
        const rows: React.ReactElement[] = [];

        for (let i = 0; i < list.length; i++) {
            const segment = list[i];
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

            const rowClassName = [
                inGroup ? "seg-group" : "",
                isGroupParentRow ? "seg-group-parent" : "",
                isGroupChildRow ? "seg-group-child" : "",
            ]
                .filter(Boolean)
                .join(" ");

            const displayAverage = showCumulativeTimes ? totalAvg + segment.average : segment.average;

            const displayPB = showCumulativeTimes ? totalPB + segment.pb : segment.pb;

            rows.push(
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

                        <td>
                            <div
                                style={{
                                    display: "flex",
                                    alignItems: "center",
                                    gap: 8,
                                }}
                            >
                                <input
                                    id={`segment-icon-${segment.id}`}
                                    type="file"
                                    accept="image/*"
                                    style={{ display: "none" }}
                                    onChange={(e) => updateSegmentIcon(segment.id, e)}
                                />

                                {!segment.icon ? (
                                    <img
                                        src={addIcon}
                                        alt="Choose image"
                                        title="Choose image"
                                        onClick={() => document.getElementById(`segment-icon-${segment.id}`)?.click()}
                                        style={{
                                            width: 24,
                                            height: 24,
                                            cursor: "pointer",
                                            backgroundColor: "#fff",
                                            border: "1px solid #666",
                                            borderRadius: 2,
                                            padding: 2,
                                            boxSizing: "border-box",
                                        }}
                                    />
                                ) : (
                                    <>
                                        <img
                                            src={segment.icon}
                                            alt=""
                                            title="Choose a different image"
                                            onClick={() =>
                                                document.getElementById(`segment-icon-${segment.id}`)?.click()
                                            }
                                            style={{
                                                width: 24,
                                                height: 24,
                                                objectFit: "contain",
                                                border: "1px solid #666",
                                                borderRadius: 2,
                                                cursor: "pointer",
                                            }}
                                        />

                                        <img
                                            src={removeIcon}
                                            alt="Clear image"
                                            title="Clear image"
                                            onClick={() =>
                                                updateSegment(segment.id, (s) => ({
                                                    ...s,
                                                    icon: "",
                                                }))
                                            }
                                            style={{
                                                width: 24,
                                                height: 24,
                                                cursor: "pointer",
                                                backgroundColor: "#fff",
                                                border: "1px solid #666",
                                                borderRadius: 2,
                                                padding: 2,
                                                boxSizing: "border-box",
                                            }}
                                        />
                                    </>
                                )}
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
                                    time={displayAverage}
                                    onChange={(newAverage) => {
                                        updateSegmentTimes(
                                            segment.id,
                                            showCumulativeTimes ? newAverage - totalAvg : newAverage,
                                            segment.pb,
                                        );
                                    }}
                                />
                            )}
                        </td>

                        <td>
                            {!hasChildren && (
                                <TimeRow
                                    time={displayPB}
                                    onChange={(newPB) => {
                                        updateSegmentTimes(
                                            segment.id,
                                            segment.average,
                                            showCumulativeTimes ? newPB - totalPB : newPB,
                                        );
                                    }}
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
                </React.Fragment>,
            );

            // Leaf segments contribute to running totals
            if (!hasChildren) {
                totalAvg += segment.average;
                totalPB += segment.pb;
            }

            // Children continue from current totals instead of restarting
            if (hasChildren) {
                const childResult = renderRows(segment.children, depth + 1, ownGroup, true, totalAvg, totalPB);

                rows.push(...childResult.rows);

                totalAvg = childResult.totalAvg;
                totalPB = childResult.totalPB;
            }
        }

        return {
            rows,
            totalAvg,
            totalPB,
        };
    }

    return (
        <div className="container form-container">
            <h2>{editing ? "Editing Split File" : "New Split File"}</h2>
            <form id="split-form" noValidate>
                <div className="row">
                    <label htmlFor="game_name">Game Name</label>
                    <div className="autocomplete" ref={gameAutocompleteRef}>
                        <input
                            value={gameName}
                            onFocus={() => setGameActive(true)}
                            onClick={() => setGameActive(true)}
                            onChange={(e) => {
                                setGameName(e.target.value);

                                // User is no longer using the selected speedrun.com game.
                                setGameID("");
                                setCategoryName("");
                                setCategoryID("");
                                setCategories([]);
                            }}
                            autoComplete="off"
                        />
                        {gameActive && games.length > 0 && (
                            <ul className="autocomplete-list" style={{ width: `${Math.ceil(longestGameWidth)}px` }}>
                                {games.map((game) => (
                                    <li
                                        key={game.id}
                                        onMouseDown={() => {
                                            selectingGame.current = true;
                                            console.log("Selected gameID: ", game.id);
                                            console.log("Selected game: ", game.name);
                                            setGameName(game.name);
                                            setGameID(game.id);

                                            if (game.platforms.length > 0) {
                                                setPlatforms(game.platforms);
                                            }

                                            setCategoryName("");
                                            setCategoryID("");

                                            setGameActive(false);
                                            // setShowGames(false);
                                        }}
                                    >
                                        {game.name}
                                    </li>
                                ))}
                            </ul>
                        )}
                    </div>
                </div>

                <div className="row">
                    <label htmlFor="game_category">Category</label>
                    <div className="autocomplete" ref={categoryAutocompleteRef}>
                        <input
                            value={categoryName}
                            onFocus={() => setCategoryActive(true)}
                            onClick={() => setCategoryActive(true)}
                            onChange={(e) => setCategoryName(e.target.value)}
                            autoComplete="off"
                        />
                        {categoryActive && categories.length > 0 && (
                            <ul className="autocomplete-list" style={{ width: `${Math.ceil(longestCategoryWidth)}px` }}>
                                {categories.map((category) => (
                                    <li
                                        key={category.id}
                                        onMouseDown={() => {
                                            setCategoryName(category.name);
                                            setCategoryID(category.id);
                                            setCategoryActive(false);
                                            // setShowCategories(false);
                                        }}
                                    >
                                        {category.name}
                                    </li>
                                ))}
                            </ul>
                        )}
                    </div>
                </div>

                <div className="row" style={{ marginTop: 10, marginBottom: 10 }}>
                    <label htmlFor="platform">Platform</label>
                    <select
                        style={{ marginLeft: 10 }}
                        id="platform"
                        value={platform}
                        onChange={(e) => setPlatform(e.target.value)}
                    >
                        {platforms.map((p) => (
                            <option key={p.id} value={p.name}>
                                {p.name}
                            </option>
                        ))}
                    </select>
                    <div className="row">
                        <label htmlFor="skin">Skin</label>

                        <select
                            style={{ marginLeft: 10 }}
                            id="skin"
                            value={selectedSkin}
                            onChange={(e) => setSelectedSkin(e.target.value)}
                        >
                            {availableSkins.map((skin) => (
                                <option key={skin} value={skin}>
                                    {skin}
                                </option>
                            ))}
                        </select>
                    </div>
                </div>

                <div className="row">
                    <label htmlFor="runattempts">Attempts</label>
                    <input
                        onChange={(e) => setAttempts(Number(e.target.value))}
                        value={attempts ?? 0}
                        id="runattempts"
                        name="attempts"
                        inputMode="numeric"
                    />
                </div>

                <div className="row">
                    <label htmlFor="offset">Start Offset (milliseconds)</label>
                    <input
                        onChange={(e) => handleOffsetChange(e.target.value)}
                        id="offsetMS"
                        name="offsetMS"
                        type="text"
                        autoComplete="off"
                        value={offsetText}
                    />
                </div>

                <div
                    style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                        marginTop: 20,
                        marginBottom: 20,
                    }}
                >
                    <button onClick={() => addSegment(null)} type="button">
                        Add Segment
                    </button>

                    <button type="button" onClick={() => setShowCumulativeTimes((v) => !v)}>
                        {showCumulativeTimes ? "Cumulative Times" : "Segment Times"}
                    </button>
                </div>

                <div className="datagrid-container">
                    <div className="datagrid">
                        {segments && segments.length > 0 && (
                            <table cellSpacing={0} className="datagrid" id="tbl-segments">
                                <thead>
                                    <tr>
                                        <th style={{ width: "5%" }}>#</th>
                                        <th style={{ width: "12%" }}>Icon</th>
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
                                <tbody>{renderRows(segments, 0, null, false, 0, 0).rows}</tbody>
                            </table>
                        )}
                    </div>
                </div>

                <hr />

                <div id="expoter" style={{ display: editing ? "block" : "none" }}>
                    <button
                        onClick={async (e) => {
                            e.preventDefault();
                            await ExportSplitFile(platform);
                        }}
                    >
                        Export Splitfile
                    </button>
                </div>

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
