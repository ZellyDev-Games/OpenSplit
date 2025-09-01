import {useEffect, useState} from "react";
import {EventsOn} from "../../wailsjs/runtime";
import {Split} from "../../wailsjs/go/session/Service";

export default function Timer() {
    const [time, setTime] = useState(0);

    useEffect(() => {
        const off = EventsOn("timer:update", (val: number) => {
           setTime(val);
        });
    }, []);

    return (<span>{formatDuration(time)} <button onClick={Split}>Split</button> </span>)
}

function formatDuration(ms: number, fmt = "HH:MM:SS.cc"): string {
    const totalSeconds = Math.floor(ms / 1000);
    const hours   = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    const millis  = ms % 1000;
    const centis  = Math.floor(millis / 10);

    return fmt
        .replace(/HH/g, hours.toString().padStart(2, "0"))
        .replace(/H/g, hours.toString())
        .replace(/MM/g, minutes.toString().padStart(2, "0"))
        .replace(/M/g, minutes.toString())
        .replace(/SS/g, seconds.toString().padStart(2, "0"))
        .replace(/S/g, seconds.toString())
        .replace(/mmm/g, millis.toString().padStart(3, "0"))
        .replace(/cc/g, centis.toString().padStart(2, "0"));
}