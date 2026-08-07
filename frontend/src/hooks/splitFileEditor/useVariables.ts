import { useEffect, useState } from "react";

import { SearchVariables } from "../../../wailsjs/go/speedrun/Service";
import { SelectedVariable, Variable } from "../../components/splitFileEditor/types/variables";
import SplitFilePayload from "../../models/splitFilePayload";

type UseVariablesProps = {
    categoryID: string;
    splitFilePayload: SplitFilePayload | null;
};

export function useVariables({ categoryID, splitFilePayload }: UseVariablesProps) {
    const [variables, setVariables] = useState<Variable[]>([]);
    const [selectedVariables, setSelectedVariables] = useState<Record<string, SelectedVariable>>({});

    useEffect(() => {
        if (!categoryID) {
            setVariables([]);
            setSelectedVariables({});
            return;
        }

        const load = async () => {
            const result = await SearchVariables(categoryID);

            setVariables(
                result.data.map((v) => ({
                    id: v.id,
                    name: v.name,
                    default: v.values.default,
                    options: Object.entries(v.values.values).map(([id, value]) => ({
                        id,
                        label: value.label,
                    })),
                })),
            );

            const values: Record<string, SelectedVariable> = {};

            (splitFilePayload?.variables ?? []).forEach((v) => {
                values[v.id] = {
                    name: v.name,
                    value: v.value,
                    label: v.label,
                };
            });

            setSelectedVariables(values);
        };

        void load();
    }, [categoryID, splitFilePayload]);

    return {
        variables,
        selectedVariables,
        setSelectedVariables,
    };
}
