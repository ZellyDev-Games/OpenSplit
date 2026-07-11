import { forwardRef, useEffect, useImperativeHandle, useState } from "react";

import { msToParts, partsToMS } from "../splitter/Timer";

type TimeRowProps = {
    time: number | null;
    onChange?: (millis: number) => void;
};

type Handle = {
    getMillis(): number;
};

export const TimeRow = forwardRef<Handle, TimeRowProps>((props, ref) => {
    const [hours, setHours] = useState("");
    const [minutes, setMinutes] = useState("");
    const [seconds, setSeconds] = useState("");
    const [centis, setCentis] = useState("");

    useEffect(() => {
        if (props.time == null) {
            setHours("");
            setMinutes("");
            setSeconds("");
            setCentis("");
            return;
        }

        const p = msToParts(props.time);

        setHours(String(p.hours));
        setMinutes(String(p.minutes));
        setSeconds(String(p.seconds));
        setCentis(String(p.centis));
    }, [props.time]);

    const emitChange = (h = hours, m = minutes, s = seconds, c = centis) => {
        props.onChange?.(
            partsToMS({
                negative: false,
                hours: parseInt(h || "0", 10),
                minutes: parseInt(m || "0", 10),
                seconds: parseInt(s || "0", 10),
                centis: parseInt(c || "0", 10),
            }),
        );
    };

    useImperativeHandle(ref, () => ({
        getMillis() {
            return partsToMS({
                negative: false,
                hours: parseInt(hours || "0", 10),
                minutes: parseInt(minutes || "0", 10),
                seconds: parseInt(seconds || "0", 10),
                centis: parseInt(centis || "0", 10),
            });
        },
    }));

    return (
        <div className="segment-time">
            <input
                value={hours}
                onChange={(e) => {
                    setHours(e.target.value);
                    emitChange(e.target.value, minutes, seconds, centis);
                }}
            />
            <span>:</span>

            <input
                value={minutes}
                onChange={(e) => {
                    setMinutes(e.target.value);
                    emitChange(hours, e.target.value, seconds, centis);
                }}
            />

            <span>:</span>

            <input
                value={seconds}
                onChange={(e) => {
                    setSeconds(e.target.value);
                    emitChange(hours, minutes, e.target.value, centis);
                }}
            />

            <span>.</span>

            <input
                value={centis}
                onChange={(e) => {
                    setCentis(e.target.value);
                    emitChange(hours, minutes, seconds, e.target.value);
                }}
            />
        </div>
    );
});
