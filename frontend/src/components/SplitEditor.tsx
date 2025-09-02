import {session} from "../../wailsjs/go/models";
import React from "react";

type SplitEditorProps = {
    splitFilePayload?: session.SplitFilePayload | null;
};

export default function SplitEditor({splitFilePayload} : SplitEditorProps) {
    const [gameName, setGameName] = React.useState<string>(splitFilePayload?.game_name ?? "");
    const [gameCategory, setGameCategory] = React.useState<string>(splitFilePayload?.game_category ?? "");
    const [segments, setSegments] = React.useState<session.SegmentPayload[]>(splitFilePayload?.segments ?? []);

    return (
        <div>
            <form id="split-form" action="/splits/save" method="post" noValidate>
                <fieldset>
                    <legend>Game Info</legend>

                    <div className="row">
                        <label htmlFor="game_name">Game Name</label>
                        <input id="game_name" name="game_name" type="text" autoComplete="off" required/>
                    </div>

                    <div className="row">
                        <label htmlFor="game_category">Category</label>
                        <input id="game_category" name="game_category" type="text" autoComplete="off" required/>
                    </div>

                    <div className="row">
                        <label htmlFor="attempts">Attempts</label>
                        <input id="attempts" name="attempts" type="number" inputMode="numeric" min="0" step="1" value="0"
                               required/>
                    </div>
                </fieldset>

                <fieldset>
                    <legend>Segments</legend>

                    <div className="toolbar">
                        <div className="left">
                            <button type="button" id="add-segment">Add Segment</button>
                        </div>
                        <div className="hint">Times are numbers (e.g., milliseconds). “ID” accepts comma-separated numbers
                            (will be parsed to number[] on save).
                        </div>
                    </div>

                    <ol id="segments" className="segments"></ol>
                </fieldset>

                <div className="actions">
                    <button type="submit" className="primary">Save</button>
                    <button type="reset">Reset</button>
                </div>
            </form>
        </div>
    )
}