import { useEffect, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";

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
    padMinutes: boolean;
    padSeconds: boolean;
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
};

export default function Timer({ offset }: TimerParams) {
    const [time, setTime] = useState(offset || 0);

    useEffect(() => {
        return EventsOn("timer:update", (val: number) => {
            setTime(val);
        });
    }, []);

    const formattedTimeParts = formatDuration(msToParts(time));

    return (
        <div className={"timer-container"}>
            <div className="time-container" aria-label="formatted duration">
                <span className="time-sign">{time < 0 && "-"}</span>
                <span className="time-hours" data-unit="hours" data-present={formattedTimeParts.showHours ? "1" : "0"}>
                    <strong>{formattedTimeParts.hoursText}</strong>
                </span>
                <span className="time-sep-hm" aria-hidden="true">
                    {formattedTimeParts.sepHM}
                </span>
                <span
                    className="time-minutes"
                    data-unit="minutes"
                    data-present={formattedTimeParts.showMinutes ? "1" : "0"}
                >
                    {formattedTimeParts.minutesText}
                </span>
                <span className="time-sep-ms" aria-hidden="true">
                    {formattedTimeParts.sepMS}
                </span>
                <span className="time-seconds" data-unit="seconds">
                    {formattedTimeParts.secondsText}
                </span>
                <span className="time-sep-sc" aria-hidden="true">
                    {formattedTimeParts.sepSC}
                </span>
                <span className="time-centis" data-unit="centis">
                    <small>{formattedTimeParts.centisText}</small>
                </span>
            </div>
        </div>
    );
}

export function msToParts(ms: number): TimeParts {
    const negative = ms < 0;
    const abs = Math.abs(ms);
    const totalSeconds = Math.floor(abs / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    const centis = Math.floor((abs % 1000) / 10);

    return {
        hours: hours,
        minutes: minutes,
        seconds: seconds,
        centis: centis,
        negative: negative,
    };
}

export function partsToMS(parts: TimeParts): number {
    const negative = parts.negative;
    let abs = 0;
    abs += parts.hours * 3600000;
    abs += parts.minutes * 60000;
    abs += parts.seconds * 1000;
    abs += parts.centis * 10;

    return negative ? abs * -1 : abs;
}

export function formatDuration(timeParts: TimeParts, showSign: boolean = false): FormattedTimeParts {
    // What to show
    const showHours = timeParts.hours > 0;
    const showMinutes = showHours ? true : timeParts.minutes > 0; // if hours>0 we show minutes (padded); else show only if minutes>0
    const padMinutes = showHours; // minutes padded when hours>0
    const padSeconds = timeParts.minutes > 0 || timeParts.hours > 0; // seconds padded when any minutes exist

    // Text values (empty string means “render span but no value”)
    const hoursText = showHours ? String(timeParts.hours) : "";
    const minutesText = showMinutes
        ? padMinutes
            ? String(timeParts.minutes).padStart(2, "0")
            : String(timeParts.minutes)
        : "";
    const secondsText = padSeconds ? String(timeParts.seconds).padStart(2, "0") : String(timeParts.seconds);
    const centisText = String(timeParts.centis).padStart(2, "0");

    // Separators only if the left side is present
    const sepHM = showHours && showMinutes ? ":" : "";
    const sepMS = showMinutes ? ":" : "";
    const sepSC = "."; // always show dot before centis

    return {
        isNegative: timeParts.negative,
        showSign: timeParts.negative || showSign,
        showHours: showHours,
        showMinutes: showMinutes,
        padMinutes: padMinutes,
        padSeconds: padSeconds,
        sepHM: sepHM,
        sepMS: sepMS,
        sepSC: sepSC,
        hoursText: hoursText,
        minutesText: minutesText,
        secondsText: secondsText,
        centisText: centisText,
    };
}

export function displayFormattedTimeParts(formattedParts: FormattedTimeParts): string[] {
    let timeString = "";
    if (formattedParts.showSign) {
        timeString = formattedParts.isNegative ? "-" : "+";
    }

    if (formattedParts.showHours) {
        timeString += formattedParts.hoursText;
    }

    if (formattedParts.showMinutes) {
        timeString += `${formattedParts.sepHM}${formattedParts.minutesText}`;
    }

    timeString += `${formattedParts.sepMS}${formattedParts.secondsText}`;
    const centisString = `${formattedParts.sepSC}${formattedParts.centisText}`;
    return [timeString, centisString];
}

export const numeric = (s: string) => /^[+-]?\d+$/.test(s);
