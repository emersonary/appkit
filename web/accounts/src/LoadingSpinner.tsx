import './loadingSpinner.css';

type LoadingSpinnerProps = {
  className?: string;
  label?: string;
  size?: 'sm' | 'md';
};

export function LoadingSpinner({ className = '', label, size = 'md' }: LoadingSpinnerProps) {
  return (
    <span
      className={['appkit-loading-spinner', `appkit-loading-spinner--${size}`, className]
        .filter(Boolean)
        .join(' ')}
      role="status"
      aria-live="polite"
    >
      <span className="appkit-loading-spinner__ring" aria-hidden="true" />
      {label ? <span className="appkit-sr-only">{label}</span> : null}
    </span>
  );
}
