import { useRef, useState } from "react";

import { useAutoCompleteWidth } from "../../../hooks/splitFileEditor/useAutoCompleteWidth";
import { useClickOutside } from "../../../hooks/splitFileEditor/useClickOutside";

export type AutocompleteProps<T> = {
    id?: string;

    value: string;

    items: T[];

    getKey: (item: T) => string;

    getLabel: (item: T) => string;

    onChange: (value: string) => void;

    onSelect: (item: T) => void;

    placeholder?: string;

    disabled?: boolean;
};

export default function Autocomplete<T>({
    id,
    value,
    items,
    getKey,
    getLabel,
    onChange,
    onSelect,
    placeholder,
    disabled = false,
}: AutocompleteProps<T>) {
    const [active, setActive] = useState(false);

    const containerRef = useRef<HTMLDivElement>(null);

    const popupWidth = useAutoCompleteWidth(items, getLabel);

    useClickOutside(containerRef, () => setActive(false));

    const selectItem = (item: T) => {
        onChange(getLabel(item));
        onSelect(item);
        setActive(false);
    };

    return (
        <div className="autocomplete" ref={containerRef}>
            <input
                id={id}
                autoComplete="off"
                disabled={disabled}
                placeholder={placeholder}
                value={value}
                onFocus={() => setActive(true)}
                onClick={() => setActive(true)}
                onChange={(e) => {
                    onChange(e.target.value);
                    setActive(true);
                }}
            />

            {active && items.length > 0 && !disabled && (
                <ul
                    className="autocomplete-list"
                    style={{
                        width: `${Math.ceil(popupWidth)}px`,
                    }}
                >
                    {items.map((item) => (
                        <li
                            key={getKey(item)}
                            onMouseDown={(e) => {
                                e.preventDefault();
                                selectItem(item);
                            }}
                        >
                            {getLabel(item)}
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
}
