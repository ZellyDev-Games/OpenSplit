import { Platform } from "../types/game";

type PlatformSelectorProps = {
    platform: string;

    platforms: Platform[];

    onChange: (value: string) => void;
};

export default function PlatformSelector({ platform, platforms, onChange }: PlatformSelectorProps) {
    return (
        <div className="row">
            <label htmlFor="platform">Platform</label>

            <select
                id="platform"
                disabled={platforms.length === 0}
                value={platform}
                onChange={(e) => onChange(e.target.value)}
            >
                {platforms.length === 0 ? (
                    <option>Loading platforms...</option>
                ) : (
                    platforms.map((item) => (
                        <option key={item.id} value={item.name}>
                            {item.name}
                        </option>
                    ))
                )}
            </select>
        </div>
    );
}
