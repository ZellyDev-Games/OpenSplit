type OffsetFieldProps = {
    value: string;

    onChange: (value: string) => void;
};

export default function OffsetField({ value, onChange }: OffsetFieldProps) {
    return (
        <div className="row">
            <label htmlFor="offsetMS">Start Offset (milliseconds)</label>

            <input
                id="offsetMS"
                name="offsetMS"
                type="text"
                autoComplete="off"
                value={value}
                onChange={(e) => onChange(e.target.value)}
            />
        </div>
    );
}
