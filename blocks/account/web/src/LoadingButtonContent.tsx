import { LoadingSpinner } from './LoadingSpinner';

type LoadingButtonContentProps = {
  label?: string;
};

export function LoadingButtonContent({ label }: LoadingButtonContentProps) {
  return (
    <>
      <LoadingSpinner size="sm" label={label} />
      {label ? <span>{label}</span> : null}
    </>
  );
}
