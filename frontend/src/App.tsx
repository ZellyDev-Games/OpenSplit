import React, { useEffect } from "react";

import { Dispatch } from "../wailsjs/go/dispatcher/Service";
import { EventsEmit, EventsOn, WindowGetPosition, WindowGetSize } from "../wailsjs/runtime";
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
    TOGGLEGLOBAL,
    FOCUS,
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
    useDetectWindowChange();
    useAppEventBindings(setViewModel);
    useWindowFocus();

    return (
        <div id="App" className="app">
            <ViewRouter model={viewModel} />
        </div>
    );
}

function useDetectWindowChange() {
    useEffect(() => {
        let lastX = 0;
        let lastY = 0;
        let lastH = 0;
        let lastW = 0;
        let init = false;

        (async () => {
            const { x, y } = await WindowGetPosition();
            lastX = x;
            lastY = y;

            const { w, h } = await WindowGetSize();
            lastW = w;
            lastH = h;
            init = true;
        })();

        const interval = window.setInterval(async () => {
            if (!init) return;
            const { x, y } = await WindowGetPosition();
            const { w, h } = await WindowGetSize();

            if (x != lastX || y != lastY || h != lastH || w != lastW) {
                console.log("window dimensions have changed, requesting save");
                lastX = x;
                lastY = y;
                lastW = w;
                lastH = h;

                EventsEmit("window:dimensions", x, y, w, h);
            }
        }, 1000);

        return () => {
            clearInterval(interval);
        };
    }, []);
}

function useAppEventBindings(setViewModel: React.Dispatch<React.SetStateAction<AppViewModel>>) {
    useEffect(() => {
        const unsubViewModel = EventsOn("ui:model", (nextModel: AppViewModel) => {
            console.log("[UI MODEL]", nextModel.view, nextModel);
            setViewModel(nextModel);
        });

        const unsubSession = EventsOn("session:update", (updatedSession: SessionPayload) => {
            setViewModel((prev) => {
                if (prev.view === AppView.Running) {
                    return {
                        ...prev,
                        session: updatedSession,
                    };
                }
                return prev;
            });
        });

        return () => {
            unsubViewModel();
            unsubSession();
        };
    }, [setViewModel]);
}

function useWindowFocus() {
    const f = async () => {
        await Dispatch(Command.FOCUS, "true");
    }

    const uf = async () => {
        await Dispatch(Command.FOCUS, "false");
    }

    useEffect(() => {
        (async () => {
            window.addEventListener("focus", f);
            window.addEventListener("blur", uf);
        })()

        return () => {
            window.removeEventListener("focus", f);
            window.removeEventListener("blur", uf);
        }
    }, []);
}
