import { Category } from "../types/game";
import MetadataAutocomplete from "./MetadataAutocomplete";

type CategorySelectorProps = {
    value: string;

    categories: Category[];

    onChange: (value: string) => void;

    onSelect: (category: Category) => void;
};

export default function CategorySelector({ value, categories, onChange, onSelect }: CategorySelectorProps) {
    return (
        <MetadataAutocomplete
            label="Category"
            htmlFor="game-category"
            value={value}
            items={categories}
            getKey={(category) => category.id}
            getLabel={(category) => category.name}
            onChange={onChange}
            onSelect={onSelect}
        />
    );
}
