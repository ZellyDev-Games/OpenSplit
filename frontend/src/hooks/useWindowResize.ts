import { useCallback, useEffect, useRef } from "react";
import { WindowGetSize, WindowSetSize } from "../../wailsjs/runtime";

export default function useWindowResize(pageKey: string) {
    const defaults = {
        edit: { w: 500, h: 900 },
        splitter: { w: 320, h: 530 },
        welcome: { w: 320, h: 530 },
    };

    const MIN_W = 100;
    const MIN_H = 100;

    const timeoutID = useRef<number | null>(null);
    function isPageKey(k: string): k is keyof typeof defaults {
        return k in defaults;
    }

    const handleResize = useCallback(() => {
        if (timeoutID.current !== null) window.clearTimeout(timeoutID.current);
        timeoutID.current = window.setTimeout(async () => {
            const appSize = await WindowGetSize();
            const h = String(Math.max(MIN_H, appSize.h));
            const w = String(Math.max(MIN_W, appSize.w));
            localStorage.setItem(`pageSize-${pageKey}-h`, h);
            localStorage.setItem(`pageSize-${pageKey}-w`, w);
            console.log(`Set page ${pageKey} to size ${w}w ${h}h`);
        }, 500);
    }, [pageKey]);

    useEffect(() => {
        (async () => {
            if (isPageKey(pageKey)) {
                const w = localStorage.getItem(`pageSize-${pageKey}-w`);
                const h = localStorage.getItem(`pageSize-${pageKey}-h`);
                const savedW = w ? parseInt(w, 10) : NaN;
                const savedH = h ? parseInt(h, 10) : NaN;
                const wFinal = !isNaN(savedW) ? savedW : defaults[pageKey].w;
                const hFinal = !isNaN(savedH) ? savedH : defaults[pageKey].h;
                console.log(`restoring page ${pageKey} to size ${wFinal}w ${hFinal}h`);
                WindowSetSize(wFinal, hFinal);
            }

            window.addEventListener("resize", handleResize);
        })();

        return () => {
            window.removeEventListener("resize", handleResize);
            if (timeoutID.current !== null) window.clearTimeout(timeoutID.current);
        };
    }, [pageKey, handleResize]);
}
