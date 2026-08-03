import { IconDefinition } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

export function IconButton({
    icon,
    onClick,
    tooltip,
    show = true,
}: {
    icon: IconDefinition;
    onClick: () => void;
    tooltip: string;
    show?: boolean;
}) {
    if (!show) return null;

    return (
        <button
            type="button"
            className="icon-btn has-tooltip"
            onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onClick();
            }}
            aria-label={tooltip}
            title={tooltip} // fallback if CSS isn't loaded
        >
            <FontAwesomeIcon icon={icon} />
            <span role="tooltip" className="tooltip-bubble">
                {tooltip}
            </span>
        </button>
    );
}
