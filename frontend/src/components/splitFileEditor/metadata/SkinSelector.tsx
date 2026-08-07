type SkinSelectorProps = {
    availableSkins: string[];

    selectedSkin: string;

    onChange: (value: string) => void;
};

export default function SkinSelector({ availableSkins, selectedSkin, onChange }: SkinSelectorProps) {
    return (
        <div className="row">
            <label htmlFor="skin">Skin</label>

            <select id="skin" value={selectedSkin} onChange={(e) => onChange(e.target.value)}>
                {availableSkins.map((skin) => (
                    <option key={skin} value={skin}>
                        {skin}
                    </option>
                ))}
            </select>
        </div>
    );
}
