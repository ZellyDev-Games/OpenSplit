export type ButtonData = {
  id: string;
  label: string;
  URL: string;
  action?: string;
  disabled?: boolean;
  title?: string;
};

type Props = {
  data: ButtonData[];
  onClick?: (item: ButtonData) => void;
  className?: string;
};

export default function ButtonList({ data, onClick, className }: Props) {
  return (
    <div className={className} role="group" aria-label="button list">
      {data.map((item) => (
        <button
          key={item.id}
          type="button"
          id={`btn-${item.id}`}
          title={item.title}
          disabled={item.disabled}
          onClick={() => !item.disabled && onClick?.(item)}
        >
          {item.label}
        </button>
      ))}
    </div>
  );
}