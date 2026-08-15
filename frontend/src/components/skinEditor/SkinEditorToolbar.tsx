import { PreviewMode } from "../../hooks/skinEditor/useSkinEditorPreview";
import { CompareAgainst, type Comparison } from "../../hooks/splitter/useComparison";
import type { SplitterLayout } from "../../hooks/splitter/useSplitterMenu";

interface Props {
    mode: PreviewMode;

    comparison: Comparison;

    layout: SplitterLayout;

    onModeChange(mode: PreviewMode): void;
    onComparisonChange(comparison: Comparison): void;
    onLayoutChange(layout: SplitterLayout): void;
    onSetDefaultLayout(): void;
}

const COMPARISONS: Comparison[] = [CompareAgainst.Average, CompareAgainst.Best, CompareAgainst.SumOfBest];

function cycleComparison(comparison: Comparison, direction: -1 | 1): Comparison {
    const index = COMPARISONS.indexOf(comparison);

    if (index < 0) {
        return COMPARISONS[0];
    }

    const nextIndex = (index + direction + COMPARISONS.length) % COMPARISONS.length;

    return COMPARISONS[nextIndex];
}

export default function SkinEditorToolbar({
    mode,
    comparison,
    layout,
    onModeChange,
    onComparisonChange,
    onLayoutChange,
    onSetDefaultLayout,
}: Props) {
    return (
        <div className="skin-preview-toolbar">
            <div className="skin-preview-toolbar-cell skin-preview-session">
                <label>Session</label>

                <select value={mode} onChange={(event) => onModeChange(event.target.value as PreviewMode)}>
                    <option value="empty">Empty</option>
                    <option value="running">In Progress</option>
                    <option value="completed">Completed</option>
                </select>
            </div>

            <div className="skin-preview-toolbar-cell skin-preview-comparison">
                <label>Comparison</label>

                <div className="row button-row">
                    <button type="button" onClick={() => onComparisonChange(cycleComparison(comparison, -1))}>
                        ◀
                    </button>

                    <span>{comparison}</span>

                    <button type="button" onClick={() => onComparisonChange(cycleComparison(comparison, 1))}>
                        ▶
                    </button>
                </div>
            </div>

            <div className="skin-preview-toolbar-cell skin-preview-layout">
                <label>Preview Layout</label>

                <div className="row radio-row">
                    <label>
                        <input
                            type="radio"
                            name="skin-preview-layout"
                            value="vertical"
                            checked={layout === "vertical"}
                            onChange={() => onLayoutChange("vertical")}
                        />
                        Vertical
                    </label>

                    <label>
                        <input
                            type="radio"
                            name="skin-preview-layout"
                            value="horizontal"
                            checked={layout === "horizontal"}
                            onChange={() => onLayoutChange("horizontal")}
                        />
                        Horizontal
                    </label>
                </div>
            </div>

            <div className="skin-preview-toolbar-cell skin-preview-default">
                <button type="button" onClick={onSetDefaultLayout}>
                    Set as Skin Default
                </button>
            </div>
        </div>
    );
}
