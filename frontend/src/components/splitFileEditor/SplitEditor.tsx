/**
 * SplitEditor is used for both:
 *
 *  - creating new split files
 *  - editing existing split files
 *
 * It manages segment hierarchy editing, Speedrun.com metadata,
 * skins, timing information, and export.
 */

import { Dispatch, ExportSplitFile } from "../../../wailsjs/go/dispatcher/Service";
import useSplitEditor from "../../hooks/splitFileEditor/useSplitEditor";
import { Command } from "../../models/command";
import SplitFilePayload from "../../models/splitFilePayload";
import { log } from "../../utils/logger";
import CategorySelector from "./metadata/CategorySelector";
import GameSelector from "./metadata/GameSelector";
import SplitMetadata from "./metadata/SplitMetadata";
import VariableSelector from "./metadata/VariableSelector";
import SegmentTable from "./segments/SegmentTable";

type SplitEditorParams = {
    splitFilePayload: SplitFilePayload | null;
};

export default function SplitEditor({ splitFilePayload }: SplitEditorParams) {
    log.debug(splitFilePayload);
    const editor = useSplitEditor(splitFilePayload);

    const {
        editing,

        games,
        selectingGame,

        gameName,
        setGameName,

        setGameID,

        categoryName,
        setCategoryName,

        setCategoryID,

        categories,

        variables,
        selectedVariables,
        setSelectedVariables,

        availableSkins,
        selectedSkin,
        setSelectedSkin,

        platform,
        setPlatform,
        platforms,
        setPlatforms,

        attempts,
        setAttempts,

        offsetText,
        handleOffsetChange,

        segments,
        setSegments,

        showCumulativeTimes,
        setShowCumulativeTimes,

        addSegment,
        updateSegment,
        deleteSegment,

        saveSplitFile,
    } = editor;

    return (
        <div className="container form-container">
            <h2>{editing ? "Editing Split File" : "New Split File"}</h2>

            <form id="split-form" noValidate>
                <GameSelector
                    value={gameName}
                    games={games}
                    selectingGame={selectingGame}
                    onChange={setGameName}
                    onClearSelection={() => {
                        setGameID("");
                        setCategoryID("");
                        setCategoryName("");
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
