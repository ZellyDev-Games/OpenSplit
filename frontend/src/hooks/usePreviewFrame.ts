import { useLayoutEffect, useRef, useState } from "react";

export function usePreviewFrame(skinCSS?: string) {
    const iframeRef = useRef<HTMLIFrameElement>(null);

    const [container, setContainer] = useState<HTMLElement | null>(null);

    useLayoutEffect(() => {
        const iframe = iframeRef.current;

        if (!iframe) {
            return;
        }

        const initialize = () => {
            const doc = iframe.contentDocument;

            if (!doc) {
                return;
            }

            doc.open();

            const skinBase = skinCSS ? skinCSS.substring(0, skinCSS.lastIndexOf("/") + 1) : "";

            doc.write(`
            <!doctype html>
            <html>
            <head>
            <meta charset="utf-8">

            ${skinBase ? `<base href="${skinBase}">` : ""}

            ${skinCSS ? `<link rel="stylesheet" href="${skinCSS}">` : ""}

            <style>
            html,
            body {
                margin:0;
                padding:0;

                overflow:visible;

                width:max-content;
                height:max-content;
            }

            #root {
                width:max-content;
                height:max-content;
            }

            #App {
                position:relative;

                overflow:visible;

                width:max-content;
                height:max-content;
            }
            </style>

            </head>

            <body>
            <div id="root">
            <div id="App"></div>
            </div>
            </body>

            </html>
            `);

            doc.close();

            setTimeout(() => {
                console.log("iframe window", {
                    innerWidth: doc.defaultView?.innerWidth,
                    innerHeight: doc.defaultView?.innerHeight,
                });

                console.log("documentElement", doc.documentElement.getBoundingClientRect());
                console.log("body", doc.body.getBoundingClientRect());
                console.log("root", doc.getElementById("root")?.getBoundingClientRect());
                console.log("app", doc.getElementById("App")?.getBoundingClientRect());
                console.log("splitter", doc.getElementById("splitter")?.getBoundingClientRect());
            });

            for (const node of Array.from(document.head.children)) {
                if (node instanceof HTMLStyleElement) {
                    const style = doc.createElement("style");

                    style.textContent = node.textContent;

                    doc.head.appendChild(style);
                }

                if (node instanceof HTMLLinkElement && node.rel === "stylesheet") {
                    const link = doc.createElement("link");

                    link.rel = "stylesheet";
                    link.href = node.href;

                    doc.head.appendChild(link);
                }
            }

            const root = doc.getElementById("App");

            if (!root) {
                return;
            }

            setContainer(root);
        };

        iframe.addEventListener("load", initialize);

        iframe.src = "about:blank";

        return () => {
            iframe.removeEventListener("load", initialize);

            setContainer(null);
        };
    }, [skinCSS]);

    return {
        iframeRef,
        container,
    };
}
