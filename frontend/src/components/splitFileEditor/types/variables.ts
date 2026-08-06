export type SelectedVariable = {
    name: string;
    value: string;
    label: string;
};

export type Variable = {
    id: string;
    name: string;
    default?: string;
    options: {
        id: string;
        label: string;
    }[];
};
