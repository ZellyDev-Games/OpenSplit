import React, { useEffect } from "react";
import { EventsOn } from "../wailsjs/runtime";

import Config from "./components/Config";
import SplitEditor from "./components/editor/SplitEditor";
import Splitter from "./components/splitter/Splitter";
import Welcome from "./components/splitter/Welcome";

import { ConfigPayload } from "./models/configPayload";
import SessionPayload from "./models/sessionPayload";
import SplitFilePayload from "./models/splitFilePayload";

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

export enum AppView {
    Welcome = "welcome",
    NewSplitFile = "new-split-file",
    EditSplitFile = "edit-split-file",
    Running = "running",
    Settings = "settings",
}

export type AppViewModel =
    | { view: AppView.Welcome }
    | { view: AppView.NewSplitFile; speedrunApiBaseUrl: string }
    | { view: AppView.EditSplitFile; splitFile: SplitFilePayload | null; speedrunApiBaseUrl: string }
    | { view: AppView.Running; session: SessionPayload }
    | { view: AppView.Settings; config: ConfigPayload };

type ViewRouterProps = { model: AppViewModel };

function ViewRouter({ model }: ViewRouterProps) {
    switch (model.view) {
        case AppView.Welcome:
            return <Welcome />;

        case AppView.NewSplitFile:
            return <SplitEditor splitFilePayload={null} speedRunAPIBase={model.speedrunApiBaseUrl} />;

        case AppView.EditSplitFile:
            return <SplitEditor splitFilePayload={model.splitFile} speedRunAPIBase={model.speedrunApiBaseUrl} />;

        case AppView.Running:
            return <Splitter sessionPayload={model.session} />;

        case AppView.Settings:
            return <Config configPayload={model.config} />;
    }
}

export default function App() {
    const [viewModel, setViewModel] = React.useState<AppViewModel>({ view: AppView.Welcome });

    useEffect(() => {
        const unsubscribe = EventsOn("ui:model", (nextModel: AppViewModel) => {
            console.log("[UI MODEL]", nextModel.view, nextModel);
            setViewModel(nextModel);
        });

        return () => unsubscribe();
    }, []);

    return (
        <div id="App" className="app">
            <ViewRouter model={viewModel} />
        </div>
    );
}
