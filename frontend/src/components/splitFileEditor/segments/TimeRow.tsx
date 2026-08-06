/**
 * TimeRow edits a duration using separate
 * HH:MM:SS.cc fields.
 *
 * Values are converted to milliseconds whenever edited.
 */

import { forwardRef } from "react";

import { TimeFieldsHandle, useTimeFields } from "../../../hooks/splitFileEditor/useTimeFields";

type TimeRowProps = {
    time: number | null;
    onChange?: (millis: number) => void;
};

export const TimeRow = forwardRef<TimeFieldsHandle, TimeRowProps>((props, ref) => {
    const { hours, minutes, seconds, centis, setHours, setMinutes, setSeconds, setCentis, emitChange } = useTimeFields(
        props,
        ref,
    );
    return (
        <div className="row segment-time">
            <input
                placeholder="H"
                value={hours}
                onChange={(e) => {
                    setHours(e.target.value);
                    emitChange(e.target.value, minutes, seconds, centis);
                }}
            />

            <span>:</span>

            <input
                placeholder="MM"
                value={minutes}
                onChange={(e) => {
                    setMinutes(e.target.value);
                    emitChange(hours, e.target.value, seconds, centis);
                }}
            />

            <span>:</span>

            <input
                placeholder="SS"
                value={seconds}
                onChange={(e) => {
                    setSeconds(e.target.value);
                    emitChange(hours, minutes, e.target.value, centis);
                }}
            />

            <span>.</span>

            <input
                placeholder="cc"
                value={centis}
                onChange={(e) => {
                    setCentis(e.target.value);
                    emitChange(hours, minutes, seconds, e.target.value);
                }}
            />
        </div>
    );
});
