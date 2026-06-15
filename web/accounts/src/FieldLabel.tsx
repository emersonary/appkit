import type { ReactNode } from 'react';

type FieldLabelProps = {
  htmlFor: string;
  className?: string;
  required?: boolean;
  requiredMarker?: string;
  icon?: ReactNode;
  children: ReactNode;
};

export function FieldLabel({
  htmlFor,
  className,
  required = false,
  requiredMarker = ' (*)',
  icon,
  children,
}: FieldLabelProps) {
  return (
    <label
      className={['appkit-field-label', className].filter(Boolean).join(' ')}
      htmlFor={htmlFor}
    >
      {icon ? <span className="appkit-field-label__icon-wrap">{icon}</span> : null}
      <span className="appkit-field-label__text">
        {children}
        {required ? requiredMarker : null}
      </span>
    </label>
  );
}
