import { Dispatch, SetStateAction } from "react";

import { SelectedVariable, Variable } from "../types/variables";
import VariableField from "./VariableField";

type VariableSelectorProps = {
    variables: Variable[];
    selectedVariables: Record<string, SelectedVariable>;
    setSelectedVariables: Dispatch<SetStateAction<Record<string, SelectedVariable>>>;
};

export default function VariableSelector({
    variables,
    selectedVariables,
    setSelectedVariables,
}: VariableSelectorProps) {
    if (variables.length === 0) {
        return null;
    }

    return (
        <>
            <label>Variables</label>

            <div className="variable-row">
                {variables.map((variable) => (
                    <VariableField
                        key={variable.id}
                        variable={variable}
                        value={selectedVariables[variable.id]}
                        onChange={(value) =>
                            setSelectedVariables((prev) => ({
                                ...prev,
                                [variable.id]: value,
                            }))
                        }
                    />
                ))}
            </div>
        </>
    );
}
