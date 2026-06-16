import { useState, type FormEvent, type ReactNode } from 'react';
import { useAccountsConfig } from './context';
import { FieldLabel } from './FieldLabel';
import { FieldIcon } from './fieldIcons';
import { LoadingSpinner } from './LoadingSpinner';
import type { AccountsClassNames, AccountsUILabels } from './types';

export type ResetPasswordPageProps = {
  token: string;
  onResetPassword: (token: string, password: string) => Promise<void>;
  classifyError?: (err: unknown) => string | null;
  onSuccess?: () => void;
  onGoToSignIn?: () => void;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
  renderBackToLogin?: (className?: string) => ReactNode;
  renderRequestNew?: (className?: string) => ReactNode;
};

export function ResetPasswordPage({
  token,
  onResetPassword,
  classifyError,
  onSuccess,
  onGoToSignIn,
  labels: labelsProp,
  classNames: classNamesProp,
  renderBackToLogin,
  renderRequestNew,
}: ResetPasswordPageProps) {
  const ctx = useAccountsConfig();
  const labels = { ...ctx.labels, ...labelsProp };
  const classNames = { ...ctx.classNames, ...classNamesProp };

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMessage(null);

    if (!token) {
      setErrorMessage(labels.resetPasswordInvalidLink);
      return;
    }
    if (password !== confirmPassword) {
      setErrorMessage(labels.passwordMismatch);
      return;
    }

    setIsSubmitting(true);
    try {
      await onResetPassword(token, password);
      setSuccess(true);
      onSuccess?.();
    } catch (err) {
      setErrorMessage(classifyError?.(err) ?? labels.genericError);
    } finally {
      setIsSubmitting(false);
    }
  }

  if (!token && !success) {
    return (
      <div className={classNames.card}>
        <p className={classNames.error} role="alert">
          {labels.resetPasswordInvalidLink}
        </p>
        <p style={{ textAlign: 'center', marginTop: '1.5rem' }}>
          {renderRequestNew?.(classNames.link) ?? renderBackToLogin?.(classNames.link) ?? null}
        </p>
      </div>
    );
  }

  if (success) {
    return (
      <div className={classNames.card}>
        <div className={classNames.success} role="status">
          {labels.resetPasswordSuccess}
        </div>
        <button
          type="button"
          className={classNames.submit}
          style={{ width: '100%', marginTop: '1rem' }}
          onClick={() => (onGoToSignIn ?? onSuccess)?.()}
        >
          {labels.goToSignIn}
        </button>
      </div>
    );
  }

  return (
    <div className={classNames.card}>
      {errorMessage && (
        <p className={classNames.error} role="alert">
          {errorMessage}
        </p>
      )}
      <form className={classNames.form} onSubmit={handleSubmit}>
        <div className={classNames.field}>
          <FieldLabel
            className={classNames.label}
            htmlFor="accounts-reset-password"
            required
            requiredMarker={labels.requiredMarker}
            icon={<FieldIcon kind="password" />}
          >
            {labels.password}
          </FieldLabel>
          <input
            id="accounts-reset-password"
            className={classNames.input}
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={6}
            autoComplete="new-password"
          />
        </div>
        <div className={classNames.field}>
          <FieldLabel
            className={classNames.label}
            htmlFor="accounts-reset-confirm-password"
            required
            requiredMarker={labels.requiredMarker}
            icon={<FieldIcon kind="confirmPassword" />}
          >
            {labels.confirmPassword}
          </FieldLabel>
          <input
            id="accounts-reset-confirm-password"
            className={classNames.input}
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            minLength={6}
            autoComplete="new-password"
          />
        </div>
        <button className={classNames.submit} type="submit" disabled={isSubmitting} aria-busy={isSubmitting}>
          {isSubmitting ? <LoadingSpinner label={labels.sending} size="sm" /> : labels.resetPasswordSubmit}
        </button>
      </form>
      <p style={{ textAlign: 'center', marginTop: '1.5rem' }}>
        {renderBackToLogin?.(classNames.link) ?? null}
      </p>
    </div>
  );
}
