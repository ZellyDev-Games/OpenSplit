import { SelectedVariable, Variable } from "../types/variables";

type VariableFieldProps = {
    variable: Variable;
    value?: SelectedVariable;
    onChange: (value: SelectedVariable) => void;
};

export default function VariableField({ variable, value, onChange }: VariableFieldProps) {
    return (
        <div className="row">
            <label>{variable.name}</label>

            <select
                id={`variable-${variable.id}`}
                value={value?.value ?? variable.default}
                onChange={(e) => {
                    const option = variable.options.find((o) => o.id === e.target.value);

                    if (!option) {
                        return;
                    }

                    onChange({
                        name: variable.name,
                        value: option.id,
                        label: option.label,
                    });
                }}
            >
                {variable.options.map((option) => (
                    <option key={option.id} value={option.id}>
                        {option.label}
                    </option>
                ))}
            </select>
        </div>
    );
}
