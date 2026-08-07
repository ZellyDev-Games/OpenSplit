type AttemptFieldProps = {
    value: number;

    onChange: (value: number) => void;
};

export default function AttemptField({ value, onChange }: AttemptFieldProps) {
    return (
        <div className="row">
            <label htmlFor="runattempts">Attempts</label>

            <input
                id="runattempts"
                name="attempts"
                inputMode="numeric"
                value={value}
                onChange={(e) => onChange(Number(e.target.value))}
            />
        </div>
    );
}
