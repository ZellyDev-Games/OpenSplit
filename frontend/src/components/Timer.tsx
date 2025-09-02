import {useEffect, useState} from "react";
import {EventsOn} from "../../wailsjs/runtime";
import {Split} from "../../wailsjs/go/session/Service";
import {Save} from "../../wailsjs/go/persister/Service";

export default function Timer() {
    const [time, setTime] = useState(0);

    useEffect(() => {
        const off = EventsOn("timer:update", (val: number) => {
            setTime(val);
        });
    }, []);

    return (
        <div className={"timer-container"}>
            {formatDuration(time)}
        </div>
    )
}

export function formatDuration(ms: number) {
    const totalSeconds = Math.floor(ms / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    const centis = Math.floor((ms % 1000) / 10);

    // What to show
    const showHours = hours > 0;
    const showMinutes = showHours ? true : minutes > 0;             // if hours>0 we show minutes (padded); else show only if minutes>0
    const padMinutes = showHours;                                    // minutes padded when hours>0
    const padSeconds = minutes > 0 || hours > 0;                     // seconds padded when any minutes exist

    // Text values (empty string means “render span but no value”)
    const hoursText = showHours ? String(hours) : "";
    const minutesText = showMinutes ? (padMinutes ? String(minutes).padStart(2, "0") : String(minutes)) : "";
    const secondsText = padSeconds ? String(seconds).padStart(2, "0") : String(seconds);
    const centisText = String(centis).padStart(2, "0");

    // Separators only if the left side is present
    const sepHM = showHours && showMinutes ? ":" : "";
    const sepMS = showMinutes ? ":" : "";
    const sepSC = "."; // always show dot before centis

    return (
        <span className="time" aria-label="formatted duration">
          <span className="time-hours" data-unit="hours" data-present={showHours ? "1" : "0"}>{hoursText}</span>
          <span className="time-sep-hm" aria-hidden="true">{sepHM}</span>
          <span className="time-minutes" data-unit="minutes" data-present={showMinutes ? "1" : "0"}>{minutesText}</span>
          <span className="time-sep-ms" aria-hidden="true">{sepMS}</span>
          <span className="time-seconds" data-unit="seconds">{secondsText}</span>
          <span className="time-sep-sc" aria-hidden="true">{sepSC}</span>
          <span className="time-centis" data-unit="centis">{centisText}</span>
        </span>
    );
}