import {session} from "../../wailsjs/go/models";
import React from "react";
import {setActiveSkin} from "../skinLoader";

type SplitEditorProps = {
    splitFilePayload?: session.SplitFilePayload | null;
};

export default function SplitEditor({splitFilePayload} : SplitEditorProps) {
    const [gameName, setGameName] = React.useState<string>(splitFilePayload?.game_name ?? "");
    const [gameCategory, setGameCategory] = React.useState<string>(splitFilePayload?.game_category ?? "");
    const [segments, setSegments] = React.useState<session.SegmentPayload[]>(splitFilePayload?.segments ?? []);

    const addSegment = () => {
        setSegments((prev) => [...prev, new session.SegmentPayload()]);
    }

    return (
        <div className="container form-container">
            <h2>Edit Split File</h2>
            <form id="split-form" action="/splits/save" method="post" noValidate>
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
                    <input id="attempts" name="attempts" inputMode="numeric" />
                </div>

                <div className="datagrid">
                    {segments.length > 0 && (
                        <table cellSpacing={0} className="datagrid" id="tbl-segments">
                            <thead>
                                <tr>
                                    <th style={{width: "5%"}}>#</th>
                                    <th style={{width: "50%"}}>Segment Name</th>
                                    <th>Average Time</th>
                                    <th>Personal Best</th>
                                    <th style={{width: "5%"}}></th>
                                </tr>
                            </thead>
                            <tbody>
                            {segments.map((segment, idx) => (
                                <tr>
                                    <td style={{textAlign: "center"}}>{idx + 1}</td>
                                    <td><input value={segment.name}/></td>
                                    <td><input value={segment.average_time}/></td>
                                    <td><input value={segment.best_time}/></td>
                                    <td></td>
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    )}
                </div>
                <div className="row">
                    <div className="actions">
                        <button onClick={addSegment} type="button" >Add Segment</button>
                    </div>
                </div>

                <hr />

                <div className="actions">
                    <button type="submit" className="primary">Save</button>
                    <button type="reset" onClick={() => setActiveSkin("default")}>Cancel</button>
                </div>
            </form>
        </div>
    )
}