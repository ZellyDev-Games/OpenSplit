import { Dispatch, SetStateAction } from "react";

import { Platform } from "../types/game";
import AttemptField from "./AttemptField";
import OffsetField from "./OffsetField";
import PlatformSelector from "./PlatformSelector";
import SkinSelector from "./SkinSelector";

type SplitMetadataProps = {
    platform: string;

    setPlatform: Dispatch<SetStateAction<string>>;

    platforms: Platform[];

    selectedSkin: string;

    setSelectedSkin: Dispatch<SetStateAction<string>>;

    availableSkins: string[];

    attempts: number;

    setAttempts: Dispatch<SetStateAction<number>>;

    offsetText: string;

    onOffsetChange: (value: string) => void;
};

export default function SplitMetadata({
    platform,
    setPlatform,
    platforms,

    selectedSkin,
    setSelectedSkin,
    availableSkins,

    attempts,
    setAttempts,

    offsetText,
    onOffsetChange,
}: SplitMetadataProps) {
    return (
        <>
            <div className="metadata-row">
                <PlatformSelector platform={platform} platforms={platforms} onChange={setPlatform} />

                <SkinSelector availableSkins={availableSkins} selectedSkin={selectedSkin} onChange={setSelectedSkin} />
            </div>

            <AttemptField value={attempts} onChange={setAttempts} />

            <OffsetField value={offsetText} onChange={onOffsetChange} />
        </>
    );
}
