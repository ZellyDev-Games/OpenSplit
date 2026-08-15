import { IconDefinition } from "@fortawesome/fontawesome-svg-core";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

type IconButtonProps = {
    icon: IconDefinition;
    tooltip?: string;
    show?: boolean;
    onClick: () => void;
};

type TooltipPosition = {
    left: number;
    top: number;
};

export function IconButton({ icon, tooltip, show = true, onClick }: IconButtonProps) {
    const buttonRef = useRef<HTMLButtonElement>(null);
    const tooltipRef = useRef<HTMLSpanElement>(null);

    const [tooltipVisible, setTooltipVisible] = useState(false);
    const [tooltipPosition, setTooltipPosition] = useState<TooltipPosition | null>(null);
    const [tooltipContainer, setTooltipContainer] = useState<HTMLElement | null>(null);

    const showTooltip = () => {
        const button = buttonRef.current;

        if (!button || !tooltip) {
            return;
        }

        const container = button.closest<HTMLElement>(".datagrid-scroll");

        if (!container) {
            return;
        }

        setTooltipContainer(container);
        setTooltipVisible(true);
    };

    const hideTooltip = () => {
        setTooltipVisible(false);
        setTooltipPosition(null);
    };

    useLayoutEffect(() => {
        if (!tooltipVisible || !tooltip || !buttonRef.current || !tooltipRef.current || !tooltipContainer) {
            return;
        }

        const button = buttonRef.current;
        const tooltipElement = tooltipRef.current;
        const scrollContainer = tooltipContainer;

        const updatePosition = () => {
            const containerRect = scrollContainer.getBoundingClientRect();
            const buttonRect = button.getBoundingClientRect();
            const tooltipRect = tooltipElement.getBoundingClientRect();

            const gap = 8;

            /*
             * Convert the button's viewport coordinates into coordinates
             * relative to the scroll container's content.
             */
            const buttonLeft = buttonRect.left - containerRect.left + scrollContainer.scrollLeft;

            const buttonTop = buttonRect.top - containerRect.top + scrollContainer.scrollTop;

            // const buttonRight = buttonLeft + buttonRect.width;

            const buttonBottom = buttonTop + buttonRect.height;

            /*
             * Visible viewport bounds expressed in scroll-container
             * content coordinates.
             */
            const minLeft = scrollContainer.scrollLeft + gap;

            const maxLeft = scrollContainer.scrollLeft + scrollContainer.clientWidth - tooltipRect.width - gap;

            const minTop = scrollContainer.scrollTop + gap;

            const maxTop = scrollContainer.scrollTop + scrollContainer.clientHeight - tooltipRect.height - gap;

            /*
             * Start centered above the button.
             */
            let left = buttonLeft + buttonRect.width / 2 - tooltipRect.width / 2;

            let top = buttonTop - tooltipRect.height - gap;

            /*
             * Horizontal collision:
             *
             * Move right if it clips left.
             * Move left if it clips right.
             */
            if (left < minLeft) {
                left = minLeft;
            } else if (left > maxLeft) {
                left = maxLeft;
            }

            /*
             * Vertical collision:
             *
             * Prefer above.
             * If that clips the top, move below.
             */
            if (top < minTop) {
                top = buttonBottom + gap;
            }

            /*
             * If below also clips the bottom, move upward as much
             * as necessary to remain inside the viewport.
             */
            if (top > maxTop) {
                top = maxTop;
            }

            /*
             * Handle very small containers / unusually large tooltips.
             */
            if (maxLeft < minLeft) {
                left = scrollContainer.scrollLeft + gap;
            } else {
                left = Math.max(minLeft, Math.min(left, maxLeft));
            }

            if (maxTop < minTop) {
                top = scrollContainer.scrollTop + gap;
            } else {
                top = Math.max(minTop, Math.min(top, maxTop));
            }

            setTooltipPosition({
                left,
                top,
            });
        };

        /*
         * The portal element must be visible long enough to measure it.
         */
        updatePosition();

        const resizeObserver = new ResizeObserver(updatePosition);

        resizeObserver.observe(scrollContainer);
        resizeObserver.observe(tooltipElement);

        window.addEventListener("resize", updatePosition);
        scrollContainer.addEventListener("scroll", updatePosition);

        return () => {
            resizeObserver.disconnect();

            window.removeEventListener("resize", updatePosition);
            scrollContainer.removeEventListener("scroll", updatePosition);
        };
    }, [tooltipVisible, tooltip, tooltipContainer]);

    if (!show) {
        return null;
    }

    const button = (
        <button
            ref={buttonRef}
            type="button"
            className="icon-btn"
            onClick={onClick}
            onMouseEnter={showTooltip}
            onMouseLeave={hideTooltip}
            onFocus={showTooltip}
            onBlur={hideTooltip}
        >
            <FontAwesomeIcon icon={icon} />
        </button>
    );

    const tooltipElement =
        tooltipVisible && tooltipContainer && tooltip
            ? createPortal(
                  <span
                      ref={tooltipRef}
                      role="tooltip"
                      className="tooltip-bubble"
                      style={
                          tooltipPosition
                              ? {
                                    left: tooltipPosition.left,
                                    top: tooltipPosition.top,
                                }
                              : {
                                    left: 0,
                                    top: 0,
                                    visibility: "hidden",
                                }
                      }
                  >
                      {tooltip}
                  </span>,
                  tooltipContainer,
              )
            : null;

    return (
        <>
            {button}
            {tooltipElement}
        </>
    );
}
