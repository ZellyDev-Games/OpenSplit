import { Dispatch, SetStateAction } from "react";

import { Comparison } from "../../../hooks/splitter/useComparison";
import { ConfigPayload } from "../../../models/configPayload";
import SessionPayload from "../../../models/sessionPayload";
import Splitter from "../../splitter/Splitter";
import PreviewTimer from "./PreviewTimer";

interface PreviewCanvasProps {
    width: number;
    height: number;

    splitterWidth: number;
    splitterHeight: number;

    paddingLeft: number;
    paddingTop: number;

    sessionPayload: SessionPayload;
    configPayload: ConfigPayload;

    disableContextMenu: boolean;
    forceExpandAll: boolean;

    comparison: Comparison;
    onComparisonChange: Dispatch<SetStateAction<Comparison>>;
}

export default function PreviewCanvas({
    width,
    height,
    splitterWidth,
    splitterHeight,
    paddingLeft,
    paddingTop,
    sessionPayload,
    configPayload,
    disableContextMenu,
    forceExpandAll,
    comparison,
    onComparisonChange,
}: PreviewCanvasProps) {
    return (
        <div
            id="preview-canvas-wrapper"
            style={{
                position: "relative",
                width,
                height,
            }}
        >
            <div
                id="preview-canvas"
                style={{
                    position: "relative",
                    width,
                    height,
                    flex: "none",
                    overflow: "visible",
                }}
            >
                <PreviewTimer value={sessionPayload.current_run?.total_time ?? 0} />

                <div
                    id="preview-splitter"
                    className="previewSplitter"
                    style={{
                        position: "absolute",
                        left: paddingLeft,
                        top: paddingTop,
                        width: splitterWidth,
                        height: splitterHeight,
                        overflow: "visible",
                    }}
                >
                    <Splitter
                        sessionPayload={sessionPayload}
                        configPayload={configPayload}
                        disableContextMenu={disableContextMenu}
                        forceExpandAll={forceExpandAll}
                        comparison={comparison}
                        onComparisonChange={onComparisonChange}
                    />
                </div>
            </div>
        </div>
    );
}
