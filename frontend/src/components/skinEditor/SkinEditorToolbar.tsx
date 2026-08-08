import { PreviewMode } from "../../hooks/skinEditor/useSkinEditorPreview";
import { CompareAgainst, type Comparison } from "../../hooks/splitter/useComparison";

interface Props {
    mode: PreviewMode;

    comparison: Comparison;

    onModeChange(mode: PreviewMode): void;

    onComparisonChange(comparison: Comparison): void;
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

export default function SkinEditorToolbar({ mode, comparison, onModeChange, onComparisonChange }: Props) {
    return (
        <div className="skin-preview-toolbar">
            <div className="row skin-preview-group">
                <label>Session</label>

                <select value={mode} onChange={(event) => onModeChange(event.target.value as PreviewMode)}>
                    <option value="empty">Empty</option>
                    <option value="running">In Progress</option>
                    <option value="completed">Completed</option>
                </select>
            </div>

            <div className="skin-preview-group">
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
        </div>
    );
}
