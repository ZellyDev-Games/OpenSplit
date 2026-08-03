/**
 * SplitEditor is used for both:
 *
 *  - creating new split files
 *  - editing existing split files
 *
 * It manages segment hierarchy editing, Speedrun.com metadata,
 * skins, timing information, and export.
 */

import React, { useEffect, useState } from "react";

import { Dispatch, ExportSplitFile } from "../../../wailsjs/go/dispatcher/Service";
import { GetAvailableSkins } from "../../../wailsjs/go/skin/Service";
import { Platforms, SearchCategories, SearchVariables } from "../../../wailsjs/go/speedrun/Service";
import { WindowCenter } from "../../../wailsjs/runtime";
import { Command } from "../../models/command";
import SegmentPayload from "../../models/segmentPayload";
import SplitFilePayload from "../../models/splitFilePayload";
import { log } from "../../utils/logger";
import CategorySelector from "./CategorySelector";
import GameSelector from "./GameSelector";
import SegmentTable from "./SegmentTable";
import { addChildRecursive, cloneSegments, deleteSegmentRecursive } from "./segmentTree";
import SplitMetadata from "./SplitMetadata";
import { Category, Platform, SelectedVariable, Variable } from "./types";
import VariableSelector from "./VariableSelector";

type SplitEditorParams = {
    splitFilePayload: SplitFilePayload | null;
};

export default function SplitEditor({ splitFilePayload }: SplitEditorParams) {
    const editing = splitFilePayload !== null;

    const [platform, setPlatform] = useState(splitFilePayload?.platform ?? "");
    const [platforms, setPlatforms] = useState<Platform[]>([]);

    const [gameName, setGameName] = useState(splitFilePayload?.game_name ?? "");
    const [gameID, setGameID] = useState(splitFilePayload?.speedrun_game_id ?? "");

    const [categoryName, setCategoryName] = useState(splitFilePayload?.game_category ?? "");
    const [categoryID, setCategoryID] = useState(splitFilePayload?.speedrun_game_category_id ?? "");
    const [categories, setCategories] = useState<Category[]>([]);

    const [variables, setVariables] = useState<Variable[]>([]);
    const [selectedVariables, setSelectedVariables] = useState<Record<string, SelectedVariable>>({});

    const [availableSkins, setAvailableSkins] = useState<string[]>([]);
    const [selectedSkin, setSelectedSkin] = useState(splitFilePayload?.selected_skin ?? "");

    const [attempts, setAttempts] = useState(splitFilePayload?.attempts ?? 0);

    const [offsetText, setOffsetText] = useState(String(splitFilePayload?.offset ?? 0));

    const [segments, setSegments] = useState<SegmentPayload[]>(cloneSegments(splitFilePayload?.segments ?? []));

    const [showCumulativeTimes, setShowCumulativeTimes] = useState(false);

    useEffect(() => {
        WindowCenter();

        log.debug("Loaded split editor", splitFilePayload?.id);

        const load = async () => {
            const [skinsResult, platformsResult] = await Promise.allSettled([GetAvailableSkins(), Platforms()]);

            if (skinsResult.status === "fulfilled") {
                setAvailableSkins(skinsResult.value);
            } else {
                log.error("Unable to load skins", skinsResult.reason);
            }

            if (platformsResult.status === "fulfilled") {
                setPlatforms(platformsResult.value);
            } else {
                log.error("Failed to load speedrun platforms", platformsResult.reason);
            }
        };

        void load();
    }, []);

    useEffect(() => {
        setSegments(cloneSegments(splitFilePayload?.segments ?? []));

        setOffsetText(String(splitFilePayload?.offset ?? 0));
    }, [splitFilePayload]);

    useEffect(() => {
        if (splitFilePayload?.selected_skin) {
            setSelectedSkin(splitFilePayload.selected_skin);
        } else if (availableSkins.length > 0) {
            setSelectedSkin(availableSkins[0]);
        }
    }, [availableSkins, splitFilePayload]);

    useEffect(() => {
        log.info("gameID changed:", gameID);
        if (!gameID) {
            setCategories([]);
            return;
        }

        const load = async () => {
            const result = await SearchCategories(gameID);
            log.debug("categories:", result);

            setCategories(result.data);
        };

        void load();
    }, [gameID]);

    useEffect(() => {
        if (!categoryID) {
            setVariables([]);
            setSelectedVariables({});
            return;
        }

        const load = async () => {
            const result = await SearchVariables(categoryID);

            setVariables(
                result.data.map((v) => ({
                    id: v.id,
                    name: v.name,
                    default: v.values.default,
                    options: Object.entries(v.values.values).map(([id, value]) => ({
                        id,
                        label: value.label,
                    })),
                })),
            );

            const values: Record<string, SelectedVariable> = {};

            (splitFilePayload?.variables ?? []).forEach((v) => {
                values[v.id] = {
                    name: v.name,
                    value: v.value,
                    label: v.label,
                };
            });

            setSelectedVariables(values);
        };

        void load();
    }, [categoryID]);

    const handleOffsetChange = (value: string) => {
        if (/^-?\d*$/.test(value)) {
            setOffsetText(value);
        }
    };

    const addSegment = (parent: SegmentPayload | null) => {
        if (parent === null) {
            setSegments((prev) => [...prev, new SegmentPayload()]);
        } else {
            setSegments((prev) => addChildRecursive(prev, parent));
        }
    };

    const updateSegment = (id: string, updater: (segment: SegmentPayload) => SegmentPayload) => {
        setSegments((prev) =>
            prev.map((segment) =>
                segment.id === id
                    ? updater(segment)
                    : {
                          ...segment,
                          children: segment.children?.map((child) => (child.id === id ? updater(child) : child)) ?? [],
                      },
            ),
        );
    };

    const deleteSegment = (id: string) => {
        setSegments((prev) => deleteSegmentRecursive(prev, id));
    };

    /**
     * Builds a SplitFile payload and submits it to the backend.
     */
    const saveSplitFile = async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();

        const payload = SplitFilePayload.createFrom({
            id: splitFilePayload?.id ?? "",

            game_name: gameName,
            speedrun_game_id: gameID,

            game_category: categoryName,
            speedrun_game_category_id: categoryID,

            variables: Object.entries(selectedVariables).map(([id, v]) => ({
                id,
                name: v.name,
                value: v.value,
                label: v.label,
            })),

            version: splitFilePayload?.version ?? 0,

            selected_skin: selectedSkin,

            segments,

            runs: splitFilePayload?.runs ?? [],

            pb: splitFilePayload?.pb ?? null,

            sob: splitFilePayload?.sob ?? 0,

            attempts,

            offset: offsetText === "" || offsetText === "-" ? 0 : Number(offsetText),

            platform,

            wr: splitFilePayload?.wr ?? {
                show: false,
                run_id: "",
                players: [],
                real_time: 0,
                in_game_time: 0,
            },

            window_x: splitFilePayload?.window_x ?? 100,
            window_y: splitFilePayload?.window_y ?? 100,
            window_width: splitFilePayload?.window_width ?? 350,
            window_height: splitFilePayload?.window_height ?? 550,
        });
        log.debug("Saving split file", {
            id: payload.id,
            segments: payload.segments.length,
            variables: payload.variables,
            game: payload.game_name,
            category: payload.game_category,
        });

        await Dispatch(Command.SUBMIT, JSON.stringify(payload));
    };

    return (
        <div className="container form-container">
            <h2>{editing ? "Editing Split File" : "New Split File"}</h2>

            <form id="split-form" noValidate>
                <GameSelector
                    value={gameName}
                    onChange={setGameName}
                    onClearSelection={() => {
                        setGameID("");
                        setCategoryID("");
                        setCategoryName("");
                        setCategories([]);
                    }}
                    onSelect={(game) => {
                        setGameID(game.id);

                        if (game.platforms.length > 0) {
                            setPlatforms(game.platforms);
                        }
                    }}
                />

                <CategorySelector
                    value={categoryName}
                    categories={categories}
                    onChange={setCategoryName}
                    onSelect={(category) => {
                        setCategoryID(category.id);
                    }}
                />

                <VariableSelector
                    variables={variables}
                    selectedVariables={selectedVariables}
                    setSelectedVariables={setSelectedVariables}
                />

                <SplitMetadata
                    platform={platform}
                    setPlatform={setPlatform}
                    platforms={platforms}
                    availableSkins={availableSkins}
                    selectedSkin={selectedSkin}
                    setSelectedSkin={setSelectedSkin}
                    attempts={attempts}
                    setAttempts={setAttempts}
                    offsetText={offsetText}
                    onOffsetChange={handleOffsetChange}
                />

                <SegmentTable
                    segments={segments}
                    setSegments={setSegments}
                    showCumulativeTimes={showCumulativeTimes}
                    setShowCumulativeTimes={setShowCumulativeTimes}
                    onAddSegment={addSegment}
                    onDeleteSegment={deleteSegment}
                    onUpdateSegment={updateSegment}
                />

                <hr />

                {editing && (
                    <div id="exporter">
                        <button
                            onClick={async (e) => {
                                e.preventDefault();

                                await ExportSplitFile(platform);
                            }}
                        >
                            Export Splitfile
                        </button>
                    </div>
                )}

                <div className="actions">
                    <button className="primary" type="submit" onClick={saveSplitFile}>
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
