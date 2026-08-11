/**
 * Live race timer.
 *
 * Receives timer updates from the backend through Wails events
 * and formats durations for display.
 */

import { useTimer } from "../../../hooks/splitter/useTimer";
import { formatDuration, msToParts } from "./timerUtils";

type TimerParams = {
    offset: number | undefined;
};

export default function Timer({ offset }: TimerParams) {
    const time = useTimer(offset);

    const formattedTimeParts = formatDuration(msToParts(time));

    return (
        <div id="time-container" className="row" aria-label="formatted duration">
            <span id="time-sign">{time < 0 && "-"}</span>
            <span id="time-hours" data-present={formattedTimeParts.showHours ? "1" : "0"}>
                <strong>{formattedTimeParts.hoursText}</strong>
            </span>
            <span id="time-sep-hm" aria-hidden="true">
                {formattedTimeParts.sepHM}
            </span>
            <span id="time-minutes" data-present={formattedTimeParts.showMinutes ? "1" : "0"}>
                {formattedTimeParts.minutesText}
            </span>
            <span id="time-sep-ms" aria-hidden="true">
                {formattedTimeParts.sepMS}
            </span>
            <span id="time-seconds">{formattedTimeParts.secondsText}</span>
            <span id="time-sep-sc" aria-hidden="true">
                {formattedTimeParts.sepSC}
            </span>
            <span id="time-centis">
                <small>{formattedTimeParts.centisText}</small>
            </span>
        </div>
    );
}
