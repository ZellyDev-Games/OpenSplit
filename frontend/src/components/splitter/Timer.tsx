import {useEffect, useState} from "react";
import {EventsOn} from "../../../wailsjs/runtime";
import useWindowResize from "../../hooks/useWindowResize";

export type TimeParts = {
    hours: number;
    minutes: number;
    seconds: number;
    centis: number;
}

export type FormattedTimeParts = {
    showHours : boolean;
    showMinutes : boolean;
    padMinutes : boolean;
    padSeconds : boolean;
    sepHM : string;
    sepMS : string;
    sepSC : string;
    hoursText : string;
    minutesText : string;
    secondsText : string;
    centisText : string;
}

export default function Timer() {
    useWindowResize("app");
    const [time, setTime] = useState(0);

    useEffect(() => {
        return EventsOn("timer:update", (val: number) => {
            setTime(val);
        })
    }, []);

    const formattedTimeParts = formatDuration(msToParts(time))

    return (
        <div className={"timer-container"}>
            <div className="time-container" aria-label="formatted duration">
                <span className="time-hours" data-unit="hours" data-present={formattedTimeParts.showHours ? "1" : "0"}>
                    <strong>{formattedTimeParts.hoursText}</strong>
                </span>
                <span className="time-sep-hm" aria-hidden="true">{formattedTimeParts.sepHM}</span>
                <span className="time-minutes" data-unit="minutes" data-present={formattedTimeParts.showMinutes ? "1" : "0"}>
                    {formattedTimeParts.minutesText}
                </span>
                <span className="time-sep-ms" aria-hidden="true">{formattedTimeParts.sepMS}</span>
                <span className="time-seconds" data-unit="seconds">{formattedTimeParts.secondsText}</span>
                <span className="time-sep-sc" aria-hidden="true">{formattedTimeParts.sepSC}</span>
                <span className="time-centis" data-unit="centis"><small>{formattedTimeParts.centisText}</small></span>
            </div>
        </div>
    )
}

export function stringToParts(time: string): TimeParts {
    const timeParts = time.split(":");
    return {
        hours: Number(timeParts[0]),
        minutes: Number(timeParts[1]),
        seconds: Number(timeParts[2].split(".")[0]),
        centis: Number(timeParts[2].split(".")[1])
    }
}


export function msToParts(ms: number) {
    const totalSeconds = Math.floor(ms / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    const centis = Math.floor((ms % 1000) / 10);

    return {
        hours: hours,
        minutes: minutes,
        seconds: seconds,
        centis: centis,
    }
}

export function formatDuration(timeParts: TimeParts) : FormattedTimeParts {
    // What to show
    const showHours = timeParts.hours > 0;
    const showMinutes = showHours ? true : timeParts.minutes > 0;             // if hours>0 we show minutes (padded); else show only if minutes>0
    const padMinutes = showHours;                                    // minutes padded when hours>0
    const padSeconds = timeParts.minutes > 0 || timeParts.hours > 0;                     // seconds padded when any minutes exist

    // Text values (empty string means “render span but no value”)
    const hoursText = showHours ? String(timeParts.hours) : "";
    const minutesText = showMinutes ? (padMinutes ? String(timeParts.minutes).padStart(2, "0") : String(timeParts.minutes)) : "";
    const secondsText = padSeconds ? String(timeParts.seconds).padStart(2, "0") : String(timeParts.seconds);
    const centisText = String(timeParts.centis).padStart(2, "0");

    // Separators only if the left side is present
    const sepHM = showHours && showMinutes ? ":" : "";
    const sepMS = showMinutes ? ":" : "";
    const sepSC = "."; // always show dot before centis

    return {
        showHours : showHours,
        showMinutes : showMinutes,
        padMinutes : padMinutes,
        padSeconds : padSeconds,
        sepHM : sepHM,
        sepMS : sepMS,
        sepSC : sepSC,
        hoursText : hoursText,
        minutesText : minutesText,
        secondsText : secondsText,
        centisText : centisText,
    }
}