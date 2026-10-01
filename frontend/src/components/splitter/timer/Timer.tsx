/**
 * Live race timer.
 *
 * Receives timer updates from the backend through Wails events
 * and formats durations for display.
 */

import { useTimer } from "../../../hooks/splitter/useTimer";
import { formatDuration, msToParts, useMillisecondDisplay } from "./timerUtils";

type TimerParams = {
    offset: number | undefined;
};

export default function Timer({ offset }: TimerParams) {
    const time = useTimer(offset);
    const showMilliseconds = useMillisecondDisplay();

    const formattedTimeParts = formatDuration(msToParts(time), false, showMilliseconds);

    return (
        <div id="time-container" className="row" aria-label="formatted duration">
            <span id="time-sign" data-present={time < 0 ? "1" : "0"}>
                {time < 0 && "-"}
            </span>
            <span id="time-hours" data-present={formattedTimeParts.showHours ? "1" : "0"}>
                <strong>{formattedTimeParts.hoursText}</strong>
            </span>
            <span id="time-sep-hm" data-present={formattedTimeParts.showHours ? "1" : "0"} aria-hidden="true">
                {formattedTimeParts.sepHM}
            </span>
            <span id="time-minutes" data-present={formattedTimeParts.showMinutes ? "1" : "0"}>
                {formattedTimeParts.minutesText}
            </span>
            <span id="time-sep-ms" data-present={formattedTimeParts.showMinutes ? "1" : "0"} aria-hidden="true">
                {formattedTimeParts.sepMS}
            </span>
            <span id="time-seconds">{formattedTimeParts.secondsText}</span>
            <span id="time-sep-sc" aria-hidden="true">
                {formattedTimeParts.sepSC}
            </span>
            <span id="time-centis">
                <small>{formattedTimeParts.fractionText}</small>
            </span>
        </div>
    );
}
