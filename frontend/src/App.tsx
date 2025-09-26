import React, { useEffect } from "react";

import { EventsOn } from "../wailsjs/runtime";
import SplitEditor from "./components/editor/SplitEditor";
import Splitter from "./components/splitter/Splitter";
import Welcome from "./components/splitter/Welcome";
import SessionPayload from "./models/sessionPayload";
import SplitFilePayload from "./models/splitFilePayload";

export enum State {
    WELCOME = 0,
    NEWFILE = 1,
    EDITING = 2,
    RUNNING = 3,
}

type Model =
    | { tag: State.WELCOME }
    | { tag: State.NEWFILE; speedRunAPIBase: string }
    | { tag: State.EDITING; splitFilePayload: SplitFilePayload | null; speedRunAPIBase: string }
    | { tag: State.RUNNING; sessionPayload: SessionPayload };

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
}

type stateEnterParams =
    | [State.WELCOME, null]
    | [State.NEWFILE, null]
    | [State.EDITING, SplitFilePayload | null]
    | [State.RUNNING, SessionPayload];

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
    }
}

function App() {
    const [model, setModel] = React.useState<Model>({ tag: State.WELCOME });

    // Subscribe to state updates from the backend
    useEffect(() => {
        console.log("mounting")
        const unsubStateUpdates = EventsOn("state:enter", (...params: stateEnterParams) => {
            switch (params[0]) {
                case State.WELCOME:
                    setModel({ tag: State.WELCOME });
                    break;
                case State.NEWFILE:
                    setModel({ tag: State.NEWFILE, speedRunAPIBase: "https://speedrun.com/api/v1" });
                    break;
                case State.EDITING:
                    setModel({
                        tag: State.EDITING,
                        splitFilePayload: params[1],
                        speedRunAPIBase: "https://speedrun.com/api/v1",
                    });
                    break;
                case State.RUNNING:
                    setModel({ tag: State.RUNNING, sessionPayload: params[1] });
            }
        });

        const unsubSessionUpdates = EventsOn("session:update", (updatedSession: SessionPayload) => {
            setModel(prev => {
                console.log("[APP]", updatedSession);
                if (prev.tag === State.RUNNING) {
                    return { tag: State.RUNNING, sessionPayload: updatedSession };
                }
                return prev
            });
        })

        return () => {
            console.log("unmounting")
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
