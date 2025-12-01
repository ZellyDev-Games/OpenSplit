import { useEffect, useRef } from "react";

import { numeric, TimeParts } from "../splitter/Timer";

type timeRowParams = {
    id: string;
    time: TimeParts | null;
    onChangeCallback: (idx: string, time: TimeParts) => void;
};

type timeRowElements = {
    hours: HTMLInputElement | null;
    minutes: HTMLInputElement | null;
    seconds: HTMLInputElement | null;
    centis: HTMLInputElement | null;
};

export default function TimeRow({ id, time, onChangeCallback }: timeRowParams) {
    // Get refs to all the parts so we can update all at once
    const timeRef = useRef<timeRowElements>({
        hours: null,
        minutes: null,
        seconds: null,
        centis: null,
    });

    // Set initial values from the passed in timeParts
    useEffect(() => {
        let el = timeRef.current.hours;
        if (el) {
            el.value = time?.hours.toString() ?? "";
        }

        el = timeRef.current.minutes;
        if (el) {
            el.value = time?.minutes.toString() ?? "";
        }

        el = timeRef.current.seconds;
        if (el) {
            el.value = time?.seconds.toString() ?? "";
        }

        el = timeRef.current.centis;
        if (el) {
            el.value = time?.centis.toString() ?? "";
        }
    }, []);

    const handleChange = () => {
        let hours = timeRef.current.hours?.value ?? "0";
        let minutes = timeRef.current.minutes?.value ?? "0";
        let seconds = timeRef.current.seconds?.value ?? "0";
        let centis = timeRef.current.centis?.value ?? "0";

        hours = numeric(hours) ? hours : "";
        minutes = numeric(minutes) ? minutes : "";
        seconds = numeric(seconds) ? seconds : "";
        centis = numeric(centis) ? centis : "";

        const hoursNum = numeric(hours.trim()) ? Number(hours) : 0;
        const minutesNum = Math.min(Math.max(numeric(minutes.trim()) ? Number(minutes) : 0, 0), 59);
        const secondsNum = Math.min(Math.max(numeric(seconds.trim()) ? Number(seconds) : 0, 0), 59);
        const centisNum = Math.min(Math.max(numeric(centis.trim()) ? Number(centis) : 0, 0), 99);

        let el = timeRef.current.hours;
        if (el) {
            el.value = hoursNum.toString() ?? "";
        }

        el = timeRef.current.minutes;
        if (el) {
            el.value = minutesNum.toString() ?? "";
        }

        el = timeRef.current.seconds;
        if (el) {
            el.value = secondsNum.toString() ?? "";
        }

        el = timeRef.current.centis;
        if (el) {
            el.value = centisNum.toString() ?? "";
        }

        onChangeCallback(id, {
            negative: false,
            hours: hoursNum,
            minutes: minutesNum,
            seconds: secondsNum,
            centis: centisNum,
        });
    };

    return (
        <div className="segment-time">
            <input
                ref={(el) => {
                    timeRef.current.hours = el;
                }}
                placeholder="H"
                onChange={handleChange}
            />
            <span>:</span>
            <input
                ref={(el) => {
                    timeRef.current.minutes = el;
                }}
                placeholder="MM"
                onChange={handleChange}
            />
            <span>:</span>
            <input
                ref={(el) => {
                    timeRef.current.seconds = el;
                }}
                placeholder="SS"
                onChange={handleChange}
            />
            <span>.</span>
            <input
                ref={(el) => {
                    timeRef.current.centis = el;
                }}
                placeholder={"cc"}
                onChange={handleChange}
            />
        </div>
    );
}
