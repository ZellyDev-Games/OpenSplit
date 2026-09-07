import { SetStateAction, useEffect, useState } from "react";

import { Dispatch } from "../wailsjs/go/dispatcher/Service";
import { GetSkinAddress } from "../wailsjs/go/skin/Service";
import {
    EventsEmit,
    EventsOn,
    WindowGetPosition,
    WindowGetSize,
    WindowSetMaxSize,
    WindowSetMinSize,
    WindowSetPosition,
    WindowSetSize,
} from "../wailsjs/runtime";
import Config from "./components/Config";
import EditSkin from "./components/skinEditor/EditSkin";
import NewSkin from "./components/skinEditor/NewSkin";
import SkinEditor from "./components/skinEditor/SkinEditor";
import SplitEditor from "./components/splitFileEditor/SplitEditor";
import Splitter from "./components/splitter/Splitter";
import Welcome from "./components/Welcome";
import { Command } from "./models/command";
import { ConfigPayload } from "./models/configPayload";
import SessionPayload from "./models/sessionPayload";
import { SkinModel } from "./models/skin/editor";
import SplitFilePayload from "./models/splitFilePayload";
import { log } from "./utils/logger";

export enum AppView {
    Welcome = "welcome",
    NewSplitFile = "new-split-file",
    EditSplitFile = "edit-split-file",
    NewSkin = "new-skin",
    EditSkin = "edit-skin",
    SkinEditor = "skin-editor",
    Running = "running",
    Settings = "settings",
}

type WindowConfig = {
    width: number;
    height: number;
    resizable: boolean;
    positioned?: boolean;
    x?: number;
    y?: number;
};

export type AppViewModel = {
    window: WindowConfig;
} & (
    | { view: AppView.Welcome }
    | {
          view: AppView.NewSplitFile;
          splitFile?: SplitFilePayload | null;
      }
    | {
          view: AppView.EditSplitFile;
          splitFile: SplitFilePayload | null;
      }
    | { view: AppView.NewSkin }
    | { view: AppView.EditSkin }
    | { view: AppView.SkinEditor }
    | { view: AppView.Running; session: SessionPayload; config: ConfigPayload }
    | { view: AppView.Settings; config: ConfigPayload }
);

type ViewRouterProps = {
    model: AppViewModel;
    skinModel: SkinModel | null;
};

function ViewRouter({ model, skinModel }: ViewRouterProps) {
    switch (model.view) {
        case AppView.Welcome:
            return <Welcome />;

        case AppView.NewSplitFile:
            return <SplitEditor splitFilePayload={model.splitFile ?? null} />;

        case AppView.EditSplitFile:
            return <SplitEditor splitFilePayload={model.splitFile} />;

        case AppView.NewSkin:
            return <NewSkin />;

        case AppView.EditSkin:
            return skinModel ? <EditSkin skins={skinModel.skins} /> : <div>Loading skins...</div>;

        case AppView.SkinEditor:
            return skinModel ? <SkinEditor model={skinModel} /> : null;

        case AppView.Running:
            return <Splitter sessionPayload={model.session} configPayload={model.config} />;

        case AppView.Settings:
            return <Config configPayload={model.config} />;

        default:
            return <Welcome />;
    }
}

export default function App() {
    const [viewModel, setViewModel] = useState<AppViewModel | null>(null);

    const [skinModel, setSkinModel] = useState<SkinModel | null>(null);

    useDetectWindowChange();
    useAppEventBindings(setViewModel);
    useWindowFocus();

    useEffect(() => {
        const unsubscribeModel = EventsOn("skin:model", (model: SkinModel) => {
            setSkinModel(model);
        });

        const unsubscribeReload = EventsOn("skin:reload", (address: string) => {
            changeSkin(address);
        });

        return () => {
            unsubscribeModel();
            unsubscribeReload();
        };
    }, []);

    useEffect(() => {
        log.info("[App] Loading initial skin");

        GetSkinAddress().then(changeSkin);

        return EventsOn("skin:update", (address: string) => {
            log.info("[App] Skin updated:", address);

            changeSkin(address);
        });
    }, []);

    return (
        <div id="App" className="app panel">
            {viewModel && <ViewRouter model={viewModel} skinModel={skinModel} />}
        </div>
    );
}

function changeSkin(address: string) {
    let link = document.getElementById("skin-css") as HTMLLinkElement | null;

    if (!link) {
        link = document.createElement("link");
        link.id = "skin-css";
        link.rel = "stylesheet";
        document.head.appendChild(link);
    }

    link.href = address;

    log.info("changed skin to ", link.href);
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
            if (!init) {
                return;
            }

            const { x, y } = await WindowGetPosition();

            const { w, h } = await WindowGetSize();

            if (x !== lastX || y !== lastY || h !== lastH || w !== lastW) {
                log.debug("[App] Window position changed", {
                    x,
                    y,
                    w,
                    h,
                });

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

function useAppEventBindings(setViewModel: React.Dispatch<SetStateAction<AppViewModel | null>>) {
    useEffect(() => {
        const unsubViewModel = EventsOn("ui:model", async (nextModel: AppViewModel) => {
            log.info("[App] UI model", nextModel);

            await applyWindowConfig(nextModel);

            const size = await WindowGetSize();

            log.info("[App] Window after resize", size);

            setViewModel(nextModel);
        });

        EventsEmit("ui:ready");

        const unsubSession = EventsOn("session:update", (updatedSession: SessionPayload) => {
            setViewModel((prev) => {
                if (!prev) {
                    return prev;
                }

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
        log.debug("[App] Window focused");

        await Dispatch(Command.FOCUS, "true");
    };

    const uf = async () => {
        log.debug("[App] Window lost focus");

        await Dispatch(Command.FOCUS, "false");
    };

    useEffect(() => {
        window.addEventListener("focus", f);

        window.addEventListener("blur", uf);

        return () => {
            window.removeEventListener("focus", f);

            window.removeEventListener("blur", uf);
        };
    }, []);
}

async function applyWindowConfig(model: AppViewModel) {
    const window = model.window;

    log.info("[App] Applying window config", window);

    WindowSetMinSize(1, 1);
    WindowSetMaxSize(10000, 10000);

    await new Promise((resolve) => setTimeout(resolve, 50));

    WindowSetSize(window.width, window.height);

    if (window.positioned && window.x !== undefined && window.y !== undefined) {
        WindowSetPosition(window.x, window.y);
    }

    if (!window.resizable) {
        await new Promise((resolve) => setTimeout(resolve, 50));

        WindowSetMinSize(window.width, window.height);

        WindowSetMaxSize(window.width, window.height);
    }
}
