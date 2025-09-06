import React, {useEffect} from "react";

type timeRowParams = {
    idx: number;
    time: string;
    onChangeCallback: (idx: number, time: string) => void;
}

export default function TimeRow({idx, time, onChangeCallback}: timeRowParams) {
    const [hours, setHours] = React.useState<string>("00");
    const [minutes, setMinutes] = React.useState<string>("00");
    const [seconds, setSeconds] = React.useState<string>("00");
    const [centis, setCentis] = React.useState<string>("00");

    useEffect(() => {
        onChangeCallback(idx, `${hours}:${minutes}:${seconds}.${centis}`)
    }, [hours, minutes, seconds, centis, idx])

    const handleChange = (val: string, clamp: number, updateFunc: (val: string) => void) => {
        if(!val) {
            val = "00";
        }

        let numVal = parseInt(val, 10)
        if(isNaN(numVal)) {
            val = "00"
            updateFunc(val);
            return;
        }

        if(clamp > 0) {
            numVal = Math.min(Math.max(numVal, 0), clamp)
        }

        updateFunc(String(numVal).padStart(2, "0"))
    }

    return (<div className="segment-time" >
        <input
            placeholder="H"
            value={hours}
            onChange={(e) => handleChange(e.target.value, 0,  setHours)}
        />
        <span>:</span>
        <input
            placeholder="MM"
            value={minutes}
            onChange={(e) => handleChange(e.target.value, 59, setMinutes)}
        />
        <span>:</span>
        <input
            placeholder="SS"
            value={seconds}
            onChange={(e) => handleChange(e.target.value, 59, setSeconds)}
        />
        <span>.</span>
        <input
            placeholder={"cc"}
            value={centis}
            onChange={(e) => handleChange(e.target.value, 99, setCentis)}
        />
    </div>)
}