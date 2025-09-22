import React, { useCallback, useEffect, useRef } from "react";

import { EventsOn, WindowGetPosition, WindowGetSize, WindowSetPosition, WindowSetSize } from "../../wailsjs/runtime";
import SessionPayload from "../models/sessionPayload";
import SplitFilePayload from "../models/splitFilePayload";

export default function useWindowResize(
    pageKey: string,
): [(key: string) => Promise<[number, number]>, (key: string) => [number, number]] {
    const [loadedSplitFile, setLoadedSplitFile] = React.useState<SplitFilePayload | null>(null);
    const defaultSizes = {
        edit: { w: 1100, h: 800 },
        splitter: { w: 320, h: 530 },
        welcome: { w: 320, h: 530 },
    };

    const MIN_W = 100;
    const MIN_H = 100;

    const timeoutID = useRef<number | null>(null);
    function isPageKey(k: string): k is keyof typeof defaultSizes {
        return k in defaultSizes;
    }

    const handleResize = useCallback(() => {
        if (timeoutID.current !== null) window.clearTimeout(timeoutID.current);
        timeoutID.current = window.setTimeout(async () => {
            const appSize = await WindowGetSize();
            const h = String(Math.max(MIN_H, appSize.h));
            const w = String(Math.max(MIN_W, appSize.w));
            localStorage.setItem(`pageSize-${pageKey}-h`, h);
            localStorage.setItem(`pageSize-${pageKey}-w`, w);
            console.log(`Set pagekey ${pageKey} to size ${w}w ${h}h`);
        }, 500);
    }, [pageKey]);

    useEffect(() => {
        (async () => {
            if (isPageKey(pageKey)) {
                if (pageKey === "splitter" && loadedSplitFile !== null) {
                    console.log(
                        `restoring size from splitfile W: ${loadedSplitFile.window_params.width} H: ${loadedSplitFile.window_params.height}`,
                    );
                    WindowSetSize(loadedSplitFile.window_params.width, loadedSplitFile.window_params.height);
                    console.log(
                        `restoring position from splitfile X: ${loadedSplitFile.window_params.x} Y: ${loadedSplitFile.window_params.y}`,
                    );
                    WindowSetPosition(loadedSplitFile.window_params.x, loadedSplitFile.window_params.y);
                } else {
                    const w = localStorage.getItem(`pageSize-${pageKey}-w`);
                    const h = localStorage.getItem(`pageSize-${pageKey}-h`);
                    const savedW = w ? parseInt(w, 10) : NaN;
                    const savedH = h ? parseInt(h, 10) : NaN;
                    const wFinal = !isNaN(savedW) ? savedW : defaultSizes[pageKey].w;
                    const hFinal = !isNaN(savedH) ? savedH : defaultSizes[pageKey].h;
                    WindowSetSize(wFinal, hFinal);
                    console.log(`restoring page ${pageKey} to size ${wFinal}w ${hFinal}h`);
                }
            }
            window.addEventListener("resize", handleResize);
        })();

        const unsubscribe = EventsOn("session:update", (s: SessionPayload) => {
            setLoadedSplitFile(s.split_file ?? null);
        });

        return () => {
            unsubscribe();
            window.removeEventListener("resize", handleResize);
            if (timeoutID.current !== null) window.clearTimeout(timeoutID.current);
        };
    }, [pageKey, handleResize, loadedSplitFile]);

    return [
        async (pageKey: string) => {
            const pos = await WindowGetPosition();
            localStorage.setItem(`pagePos-${pageKey}-x`, pos.x.toString());
            localStorage.setItem(`pagePos-${pageKey}-y`, pos.y.toString());
            console.log("set window position: ", pos.x.toString(), pos.y.toString());
            return [pos.x, pos.y];
        },
        (pageKey: string): [number, number] => {
            const w = Math.max(200, Number(localStorage.getItem(`pageSize-${pageKey}-w`)));
            const h = Math.max(200, Number(localStorage.getItem(`pageSize-${pageKey}-h`)));
            console.log("get page size: ", w, h);
            return [w, h];
        },
    ];
}
