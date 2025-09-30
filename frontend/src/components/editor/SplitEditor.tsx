import { faTrash } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import React, { useEffect, useRef, useState } from "react";

import { Dispatch } from "../../../wailsjs/go/dispatcher/Service";
import { WindowCenter, WindowSetSize } from "../../../wailsjs/runtime";
import { Command } from "../../App";
import { useClickOutside } from "../../hooks/useClickOutside";
import SegmentPayload from "../../models/segmentPayload";
import SplitFilePayload from "../../models/splitFilePayload";
import { msToParts, partsToMS, TimeParts } from "../splitter/Timer";
import TimeRow from "./TimeRow";

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

export default function SplitEditor({ splitFilePayload, speedRunAPIBase }: SplitEditorParams) {
    // Clear modal results which will close the modal when we click outside the modal
    const clickOutsideRef = useRef<HTMLDivElement | null>(null);
    useClickOutside(clickOutsideRef, () => {
        setGameResults([]);
    });

    // Segment stats
    const [splitFileLoaded] = useState<boolean>(false);
    const [gameName, setGameName] = React.useState<string>("");
    const [gameCategory, setGameCategory] = React.useState<string>("");
    const [segments, setSegments] = React.useState<SegmentPayload[]>([]);
    const [attempts, setAttempts] = React.useState<number>(0);

    // Speedrun search
    const [gameResults, setGameResults] = React.useState<Game[]>([]);
    const timeoutID = useRef<number>(0);

    // Position and size the edit window
    useEffect(() => {
        WindowSetSize(1000, 900);
        WindowCenter();
    }, []);

    // Pull apart the segment times from the split file in a way our UI can use them.
    useEffect(() => {
        (async () => {
            if (!splitFilePayload) return;
            setGameName(splitFilePayload.game_name);
            setGameCategory(splitFilePayload.game_category);
            setAttempts(splitFilePayload.attempts);
            setSegments(splitFilePayload.segments);
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

    const addSegment = () => {
        setSegments((prev) => [...prev, new SegmentPayload()]);
    };

    const updateSegmentName = (idx: number, name: string) => {
        setSegments((prev) =>
            prev.map((s, i) => {
                if (idx === i) {
                    const newSegment = new SegmentPayload();
                    newSegment.id = s.id;
                    newSegment.average = s.average;
                    newSegment.pb = s.pb;
                    newSegment.gold = s.gold;
                    newSegment.name = name;
                    return newSegment;
                }
                return s;
            }),
        );
    };

    const deleteSegment = (idx: number) => {
        setSegments((prev) => prev.filter((_, i) => i !== idx));
    };

    const saveSplitFile = async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();
        const segmentPayloads = segments.map((s) => SegmentPayload.createFrom(s));
        const newSpiltFilePayload = SplitFilePayload.createFrom({
            id: splitFilePayload?.id ?? "",
            version: splitFilePayload?.version ?? 0,
            runs: splitFilePayload?.runs ?? [],
            window_x: splitFilePayload?.window_x ?? 100,
            window_y: splitFilePayload?.window_y ?? 100,
            window_height: splitFilePayload?.window_height ?? 550,
            window_width: splitFilePayload?.window_width ?? 350,
            game_name: gameName,
            game_category: gameCategory,
            segments: segmentPayloads,
            attempts: Number(attempts),
            pb: splitFilePayload?.pb ?? null,
            sob: splitFilePayload?.sob ?? 0,
        });

        const payload = JSON.stringify(newSpiltFilePayload);
        console.log(payload);
        await Dispatch(Command.SUBMIT, payload);
    };

    const handleTimeChange = (idx: number, time: TimeParts, isBest: boolean) => {
        const ms = partsToMS(time);
        const newSegments = [];
        for (let i = 0; i < segments.length; i++) {
            if (idx != i) {
                newSegments.push({ ...segments[i] });
            } else {
                const s = { ...segments[i] };
                const pb = isBest ? ms : s.pb;
                const avg = isBest ? s.average : ms;
                newSegments.push(
                    SegmentPayload.createFrom({
                        id: s.id,
                        name: s.name,
                        gold: s.gold,
                        pb: pb,
                        average: avg,
                    }),
                );
            }
        }

        setSegments(newSegments);
    };

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

                <div style={{ marginTop: 20, marginBottom: 20 }} className="row">
                    <div>
                        <button onClick={addSegment} type="button">
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
                                        <th style={{ width: "5%" }}></th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {segments.map((segment, idx) => (
                                        <tr key={idx}>
                                            <td style={{ textAlign: "center" }}>{idx + 1}</td>
                                            <td>
                                                <input
                                                    onChange={(e) => updateSegmentName(idx, e.target.value)}
                                                    value={segment.name}
                                                />
                                            </td>
                                            <td>
                                                <TimeRow
                                                    idx={idx}
                                                    time={segment.average ? msToParts(segment.average) : null}
                                                    onChangeCallback={(idx, ts) => handleTimeChange(idx, ts, false)}
                                                />
                                            </td>
                                            <td>
                                                <TimeRow
                                                    idx={idx}
                                                    time={segment.pb ? msToParts(segment.pb) : null}
                                                    onChangeCallback={(idx, ts) => handleTimeChange(idx, ts, true)}
                                                />
                                            </td>
                                            <td style={{ textAlign: "center" }}>
                                                <div onClick={() => deleteSegment(idx)}>
                                                    <FontAwesomeIcon icon={faTrash} />
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
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
