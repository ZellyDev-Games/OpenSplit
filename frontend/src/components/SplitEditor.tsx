import {session} from "../../wailsjs/go/models";
import React, {useEffect, useRef} from "react";
import {GetConfig, UpdateSplitFile} from "../../wailsjs/go/session/Service";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faTrash} from "@fortawesome/free-solid-svg-icons";
import SegmentPayload = session.SegmentPayload;
import SplitFilePayload = session.SplitFilePayload;
import TimeRow from "./TimeRow";
import {useNavigate} from "react-router";
import {useClickOutside} from "../hooks/useClickOutside";
import {WindowSetSize} from "../../wailsjs/runtime";

type SplitEditorProps = {
    loadedSplitFile : SplitFilePayload | null;
}

type Game = {
    id: string;
    names: { international: string };
    assets: { "cover-tiny": {uri: string} };
    released: string;
};

type Segment = {
    id: string;
    name: string;
    average_time: string
    best_time: string
}

export default function SplitEditor({loadedSplitFile} : SplitEditorProps) {
    const clickOutsideRef = useRef<HTMLDivElement | null>(null);
    const navigate = useNavigate();
    const [gameName, setGameName] = React.useState<string>("");
    const [gameCategory, setGameCategory] = React.useState<string>("");
    const [segments, setSegments] = React.useState<Segment[]>([]);
    const [attempts, setAttempts] = React.useState<number>(0);
    const [speedrunAPIBase, setSpeedrunAPIBase] = React.useState<string>("");
    const [gameResults, setGameResults] = React.useState<Game[]>([])
    const timeoutID = useRef<number>(0);
    useClickOutside(clickOutsideRef, () => { console.log("Click outside handler fired"); setGameResults([])});

    useEffect(() => WindowSetSize(1000, 900), [])

    useEffect(() => {
        (async() => {
            if (loadedSplitFile === null) return;
            setGameName(loadedSplitFile.game_name);
            setGameCategory(loadedSplitFile.game_category);
            setAttempts(loadedSplitFile.attempts);
            setSegments(
                loadedSplitFile ?
                    loadedSplitFile?.segments.map((s, i) => {
                    return {...s, "idx": i, average_time: s.average_time, best_time: s.best_time }
                }) : []
            )
        })()
    }, [loadedSplitFile]);

    useEffect(() => {
        (async() => {
            const config = await GetConfig()
            setSpeedrunAPIBase(config.speed_run_API_base);
        })()
    }, []);

    const searchSpeedrun = async () => {
        if(!speedrunAPIBase) return;
        const q = gameName.trim()
        if(!q) {setGameResults([]); return;}

        clearTimeout(timeoutID.current);
        const controller = new AbortController();
        timeoutID.current = setTimeout(async () => {
            fetch(`${speedrunAPIBase}/games?name=${encodeURIComponent(gameName)}`, {
                signal: controller.signal,
            })
                .then((res) => res.json())
                .then((data) => setGameResults(data.data))
                .catch((err) => {
                    if (err.name !== "AbortError") console.error("search failed:", err);
                })
        }, 500);
    }

    const addSegment = () => {
        setSegments((prev) => [...prev, {
            average_time: "0",
            best_time: "0",
            id: "",
            name: ""
        }]);
    }

    const updateSegments = (idx: number, attrName: string, attrVal: string) => {
        setSegments(prev =>
            prev.map((s, i) =>
                i === idx ? {...s, [attrName]: attrVal} : s
            )
        );
    }

    const deleteSegment = (idx: number)=> {
        setSegments(prev => prev.filter((_, i) => i !== idx));
    }

    const saveSplitFile = async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();
        const segmentPayloads = segments.map((s) => SegmentPayload.createFrom(s))
        const splitFilePayload = SplitFilePayload.createFrom({
            game_name: gameName,
            game_category: gameCategory,
            segments: segmentPayloads,
            attempts: Number(attempts)
        })

        UpdateSplitFile(splitFilePayload).then(_ => navigate("/")).catch((err) => console.log(err));
    }

    const handleTimeChange = (idx: number, time: string, isBest: boolean) => {
        setSegments(prev => prev.map((s, i) => i === idx ? {
            ...s,
            best_time: isBest ? time : s.best_time,
            average_time: isBest ? s.average_time : time,
        } : s))
    };

    return (
        <div className="container form-container" >
            <h2>Edit Split File</h2>
            <form id="split-form" noValidate>
                <div className="row">
                    <label htmlFor="game_name">Game Name</label>
                    <input value={gameName}
                           onChange={(e) => setGameName(e.target.value)}
                           onBlur={(e) => {
                               clearTimeout(timeoutID.current);
                           }}
                           onKeyUp={searchSpeedrun}
                           id="game_name"
                           name="game_name"
                           type="text"
                           autoComplete="off" />
                </div>

                {(gameResults.length > 0) &&
                <div ref={clickOutsideRef} className="autocomplete">
                    <ul>
                        {gameResults.map((gameResult) => (
                            <li onClick={() => {
                                setGameName(gameResult.names.international);
                                setGameResults([])
                            }} key={gameResult.id}>
                                <div className="autocomplete-item">
                                    <img src={gameResult.assets["cover-tiny"].uri} alt={gameResult.assets["cover-tiny"].uri} />
                                    <div className="game-info">
                                        <strong>{gameResult.names.international}</strong>
                                        <span>{gameResult.released}</span>
                                    </div>
                                </div>
                            </li>
                        ))}
                    </ul>
                </div>
                }

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
                    <input onChange={(e) => setAttempts(Number(e.target.value))} value={attempts ?? 0} id="attempts" name="attempts" inputMode="numeric" />
                </div>

                <div style={{marginTop:20, marginBottom:20}} className="row">
                    <div>
                        <button onClick={addSegment} type="button" >Add Segment</button>
                    </div>
                </div>
                <div className="datagrid-container">
                    <div className="datagrid">
                        {segments.length > 0 && (
                            <table cellSpacing={0} className="datagrid" id="tbl-segments">
                                <thead>
                                    <tr>
                                        <th style={{width: "5%"}}>#</th>
                                        <th style={{width: "50%"}}>Segment Name</th>
                                        <th>Average Time <small>(HH:MM:SS.ccc)</small></th>
                                        <th>Personal Best <small>(HH:MM:SS.ccc)</small></th>
                                        <th style={{width: "5%"}}></th>
                                    </tr>
                                </thead>
                                <tbody>
                                {segments.map((segment, idx) => (
                                    <tr key={idx}>
                                        <td style={{textAlign: "center"}}>{idx + 1}</td>
                                        <td>
                                            <input
                                                onChange={(e) => updateSegments(idx, "name", e.target.value)}
                                                value={segment.name ?? ""}
                                            />
                                        </td>
                                        <td>
                                            <TimeRow idx={idx} time={segment.average_time} onChangeCallback={(idx, ts) => handleTimeChange(idx, ts, false)} />
                                        </td>
                                        <td>
                                            <TimeRow idx={idx} time={segment.best_time} onChangeCallback={(idx, ts) => handleTimeChange(idx, ts, true)} />
                                        </td>
                                        <td style={{textAlign: "center"}}>
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
                    <button onClick={saveSplitFile} type="submit" className="primary">Save</button>
                    <button onClick={() => navigate("/")}>Cancel</button>
                </div>
            </form>
        </div>
    )
}