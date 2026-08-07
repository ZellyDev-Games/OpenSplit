import Autocomplete, { AutocompleteProps } from "../shared/AutoComplete";

type MetadataAutocompleteProps<T> = AutocompleteProps<T> & {
    label: string;
    htmlFor?: string;
};

export default function MetadataAutocomplete<T>({ label, htmlFor, ...props }: MetadataAutocompleteProps<T>) {
    return (
        <div className="row">
            <label htmlFor={htmlFor}>{label}</label>

            <Autocomplete {...props} />
        </div>
    );
}
