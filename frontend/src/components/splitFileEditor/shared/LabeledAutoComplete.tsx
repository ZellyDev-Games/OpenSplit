import Autocomplete from "./AutoComplete";

type LabeledAutocompleteProps<T> = {
    id: string;

    label: string;

    value: string;

    items: T[];

    getKey: (item: T) => string;

    getLabel: (item: T) => string;

    onChange: (value: string) => void;

    onSelect: (item: T) => void;
};

export default function LabeledAutocomplete<T>({
    id,
    label,
    value,
    items,
    getKey,
    getLabel,
    onChange,
    onSelect,
}: LabeledAutocompleteProps<T>) {
    return (
        <div className="row">
            <label htmlFor={id}>{label}</label>

            <Autocomplete
                id={id}
                value={value}
                items={items}
                getKey={getKey}
                getLabel={getLabel}
                onChange={onChange}
                onSelect={onSelect}
            />
        </div>
    );
}
