import { useState, type FormEvent, type ReactNode } from 'react';
import { useAccountsConfig } from './context';
import { FieldLabel } from './FieldLabel';
import { FieldIcon } from './fieldIcons';
import { LoadingButtonContent } from './LoadingButtonContent';
import type { AccountsClassNames, AccountsUILabels } from './types';

export type ForgotPasswordPageProps = {
  onRequestReset: (email: string) => Promise<void>;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
  renderBackToLogin?: (className?: string) => ReactNode;
};

export function ForgotPasswordPage({
  onRequestReset,
  labels: labelsProp,
  classNames: classNamesProp,
  renderBackToLogin,
}: ForgotPasswordPageProps) {
  const ctx = useAccountsConfig();
  const labels = { ...ctx.labels, ...labelsProp };
  const classNames = { ...ctx.classNames, ...classNamesProp };

  const [email, setEmail] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [sent, setSent] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setIsSending(true);
    setErrorMessage(null);
    try {
      await onRequestReset(email.trim());
      setSent(true);
    } catch {
      setErrorMessage(labels.genericError);
    } finally {
      setIsSending(false);
    }
  }

  if (sent) {
    return (
      <div className={classNames.card}>
        <div className={classNames.success} role="status">
          {labels.forgotPasswordSent}
        </div>
        <p className={classNames.muted}>{labels.forgotPasswordCheckSpam}</p>
        <p style={{ textAlign: 'center', marginTop: '1.5rem' }}>
          {renderBackToLogin?.(classNames.link) ?? null}
        </p>
      </div>
    );
  }

  return (
    <div className={classNames.card}>
      <p className={classNames.instructions ?? classNames.muted}>{labels.forgotPasswordInstructions}</p>
      {errorMessage && (
        <p className={classNames.error} role="alert">
          {errorMessage}
        </p>
      )}
      <form className={classNames.form} onSubmit={handleSubmit}>
        <div className={classNames.field}>
          <FieldLabel
            className={classNames.label}
            htmlFor="accounts-forgot-email"
            required
            requiredMarker={labels.requiredMarker}
            icon={<FieldIcon kind="email" />}
          >
            {labels.email}
          </FieldLabel>
          <input
            id="accounts-forgot-email"
            className={classNames.input}
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
        </div>
        <button className={classNames.submit} type="submit" disabled={isSending} aria-busy={isSending}>
          {isSending ? <LoadingButtonContent label={labels.sending} /> : labels.forgotPasswordSubmit}
        </button>
      </form>
      <p style={{ textAlign: 'center', marginTop: '1.5rem' }}>
        {renderBackToLogin?.(classNames.link) ?? null}
      </p>
    </div>
  );
}
