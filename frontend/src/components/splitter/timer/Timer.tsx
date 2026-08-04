/**
 * Live race timer.
 *
 * Receives timer updates from the backend through Wails events
 * and formats durations for display.
 */

import { useTimer } from "../../../hooks/splitter/useTimer";
import WorldRecord from "../../../models/worldRecord";
import { displayFormattedTimeParts, formatDuration, msToParts } from "./timerUtils";

export type TimeParts = {
    negative: boolean;
    hours: number;
    minutes: number;
    seconds: number;
    centis: number;
};

export type FormattedTimeParts = {
    isNegative: boolean;
    showSign: boolean;
    showHours: boolean;
    showMinutes: boolean;
    sepHM: string;
    sepMS: string;
    sepSC: string;
    hoursText: string;
    minutesText: string;
    secondsText: string;
    centisText: string;
};

type TimerParams = {
    offset: number | undefined;
    wr?: WorldRecord;
};

export default function Timer({ offset, wr }: TimerParams) {
    const time = useTimer(offset);

    const formattedTimeParts = formatDuration(msToParts(time));

    const rt = wr && displayFormattedTimeParts(formatDuration(msToParts(wr.real_time * 1000)));
    const igt = wr && displayFormattedTimeParts(formatDuration(msToParts(wr.in_game_time * 1000)));
    const players = wr?.players?.length ? wr.players.join(", ") : "Unknown";

    return (
        <div id="timer-container">
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
            {wr?.show && (
                <div id="world-record">
                    <div>
                        <strong>WR</strong> {players}
                    </div>

                    <div>
                        RT {rt![0]}
                        <small>{rt![1]}</small>
                    </div>

                    {wr.in_game_time > 0 && (
                        <div>
                            IGT {igt![0]}
                            <small>{igt![1]}</small>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
