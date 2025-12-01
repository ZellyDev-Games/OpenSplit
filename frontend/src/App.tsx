import React, { useEffect } from "react";

import { EventsOn } from "../wailsjs/runtime";
import Config from "./components/Config";
import SplitEditor from "./components/editor/SplitEditor";
import Splitter from "./components/splitter/Splitter";
import Welcome from "./components/splitter/Welcome";
import { ConfigPayload } from "./models/configPayload";
import SessionPayload from "./models/sessionPayload";
import SplitFilePayload from "./models/splitFilePayload";

export enum State {
    WELCOME = 0,
    NEWFILE = 1,
    EDITING = 2,
    RUNNING = 3,
    CONFIG = 4,
}

type Model =
    | { tag: State.WELCOME }
    | { tag: State.NEWFILE; speedRunAPIBase: string }
    | { tag: State.EDITING; splitFilePayload: SplitFilePayload | null; speedRunAPIBase: string }
    | { tag: State.RUNNING; sessionPayload: SessionPayload }
    | { tag: State.CONFIG; configPayload: ConfigPayload };

export enum Command {
    QUIT,
    NEW,
    LOAD,
    EDIT,
    CANCEL,
    SUBMIT,
    CLOSE,
    RESET,
    SAVE,
    SPLIT,
    UNDO,
    SKIP,
    PAUSE,
}

type stateEnterParams =
    | [State.WELCOME, null]
    | [State.NEWFILE, null]
    | [State.EDITING, SplitFilePayload | null]
    | [State.RUNNING, SessionPayload]
    | [State.CONFIG, ConfigPayload];

type StateViewProps = { model: Model };
function StateRouter({ model }: StateViewProps) {
    switch (model.tag) {
        case State.WELCOME:
            return <Welcome />;
        case State.NEWFILE:
            return <SplitEditor splitFilePayload={null} speedRunAPIBase={model.speedRunAPIBase} />;
        case State.EDITING:
            return <SplitEditor splitFilePayload={model.splitFilePayload} speedRunAPIBase={model.speedRunAPIBase} />;
        case State.RUNNING:
            return <Splitter sessionPayload={model.sessionPayload} />;
        case State.CONFIG:
            return <Config configPayload={model.configPayload} />;
    }
}

function App() {
    const [model, setModel] = React.useState<Model>({ tag: State.WELCOME });

    // Subscribe to state updates from the backend
    useEffect(() => {
        const unsubStateUpdates = EventsOn("state:enter", (...params: stateEnterParams) => {
            switch (params[0]) {
                case State.WELCOME:
                    setModel({ tag: State.WELCOME });
                    break;
                case State.NEWFILE:
                    setModel({ tag: State.NEWFILE, speedRunAPIBase: "https://www.speedrun.com/api/v1" });
                    break;
                case State.EDITING:
                    setModel({
                        tag: State.EDITING,
                        splitFilePayload: params[1],
                        speedRunAPIBase: "https://www.speedrun.com/api/v1",
                    });
                    break;
                case State.RUNNING:
                    console.log("[FSM:RUNNING]", params[1]);
                    setModel({ tag: State.RUNNING, sessionPayload: params[1] });
                    break;
                case State.CONFIG:
                    console.log(params[1]);
                    setModel({ tag: State.CONFIG, configPayload: params[1] });
            }
        });

        const unsubSessionUpdates = EventsOn("session:update", (updatedSession: SessionPayload) => {
            setModel((prev) => {
                if (prev.tag === State.RUNNING) {
                    return { tag: State.RUNNING, sessionPayload: updatedSession };
                }
                return prev;
            });
        });

        return () => {
            unsubStateUpdates();
            unsubSessionUpdates();
        };
    }, []);

    return (
        <div id="App" className="app">
            <StateRouter model={model} />
        </div>
    );
}

export default App;
