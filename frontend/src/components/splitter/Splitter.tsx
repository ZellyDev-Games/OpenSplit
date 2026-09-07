import { Dispatch, SetStateAction, useEffect, useRef } from "react";

import { Comparison, useComparison } from "../../hooks/splitter/useComparison";
import { useSegmentList } from "../../hooks/splitter/useSegmentList";
import { useSplitterLayout } from "../../hooks/splitter/useSplitterLayout";
import { SplitterLayout, useSplitterMenu } from "../../hooks/splitter/useSplitterMenu";
import { useSplitterMinimumSize } from "../../hooks/splitter/useSplitterMinimumSize";
import { useContextMenu } from "../../hooks/useContextMenu";
import { ConfigPayload } from "../../models/configPayload";
import SessionPayload from "../../models/sessionPayload";
import { ContextMenu } from "../ContextMenu";
import SplitterContent from "./SplitterContent";

type SplitterParams = {
    sessionPayload: SessionPayload;
    configPayload: ConfigPayload;
    disableContextMenu?: boolean;
    forceExpandAll?: boolean;

    comparison?: Comparison;
    onComparisonChange?: Dispatch<SetStateAction<Comparison>>;

    layout?: SplitterLayout;
    onLayoutChange?: Dispatch<SetStateAction<SplitterLayout>>;
};

export default function Splitter({
    sessionPayload,
    configPayload,
    disableContextMenu = false,
    forceExpandAll = false,
    comparison: controlledComparison,
    onComparisonChange,
    layout: controlledLayout,
    onLayoutChange,
}: SplitterParams) {
    const splitterRef = useRef<HTMLDivElement>(null);

    const contextMenu = useContextMenu();

    const { comparison, setComparison } = useComparison(controlledComparison, onComparisonChange);

    const { initialLayout } = useSplitterLayout({
        splitterRef,
        sessionPayload,
        controlledLayout,
    });

    useSplitterMinimumSize(splitterRef);

    const { layout: menuLayout, items: contextMenuItems } = useSplitterMenu({
        disableContextMenu,
        globalHotkeysInitial: configPayload.global_hotkeys_active,
        comparison,
        setComparison,
        sessionPayload,
        initialLayout,
    });

    const layout = controlledLayout ?? menuLayout;

    useEffect(() => {
        if (controlledLayout === undefined) {
            return;
        }

        if (menuLayout !== controlledLayout) {
            onLayoutChange?.(controlledLayout);
        }
    }, [controlledLayout, menuLayout, onLayoutChange]);

    const { completeClassName, rows, finalRow, containerRef } = useSegmentList({
        sessionPayload,
        comparison,
        forceExpandAll,
    });

    return (
        <div ref={splitterRef} {...(!disableContextMenu ? contextMenu.bind : {})} id="splitter" data-layout={layout}>
            {!disableContextMenu && (
                <ContextMenu state={contextMenu.state} close={contextMenu.close} items={contextMenuItems} />
            )}

            <SplitterContent
                sessionPayload={sessionPayload}
                comparison={comparison}
                forceExpandAll={forceExpandAll}
                completeClassName={completeClassName}
                containerRef={containerRef}
                rows={rows}
                finalRow={finalRow}
            />
        </div>
    );
}
