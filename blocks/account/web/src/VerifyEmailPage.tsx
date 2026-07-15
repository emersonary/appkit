import { useEffect, useState, type ReactNode } from 'react';
import { useAccountsConfig } from './context';
import { LoadingButtonContent } from './LoadingButtonContent';
import type { AccountsClassNames, AccountsUILabels } from './types';

export type VerifyEmailPageProps = {
  email: string;
  verifiedStatus?: string | null;
  onFetchStatus: (email: string) => Promise<boolean>;
  onResend: (email: string) => Promise<void>;
  onGoToSignIn?: () => void;
  onVerifiedChange?: (verified: boolean) => void;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
  renderBackToSignIn?: (className?: string) => ReactNode;
  pollIntervalMs?: number;
};

export function VerifyEmailPage({
  email,
  verifiedStatus = null,
  onFetchStatus,
  onResend,
  onGoToSignIn,
  onVerifiedChange,
  labels: labelsProp,
  classNames: classNamesProp,
  renderBackToSignIn,
  pollIntervalMs = 4000,
}: VerifyEmailPageProps) {
  const ctx = useAccountsConfig();
  const labels = { ...ctx.labels, ...labelsProp };
  const classNames = { ...ctx.classNames, ...classNamesProp };

  const [isVerified, setIsVerified] = useState(
    verifiedStatus === 'verified' || verifiedStatus === 'already',
  );
  const [isSending, setIsSending] = useState(false);
  const [resent, setResent] = useState(false);
  const [error, setError] = useState(false);

  useEffect(() => {
    if (isVerified || !email) return;

    let cancelled = false;

    async function poll() {
      try {
        const verified = await onFetchStatus(email);
        if (!cancelled && verified) {
          setIsVerified(true);
          onVerifiedChange?.(true);
        }
      } catch {
        /* ignore polling errors */
      }
    }

    poll();
    const interval = window.setInterval(poll, pollIntervalMs);
    return () => {
      cancelled = true;
      window.clearInterval(interval);
    };
  }, [email, isVerified, onFetchStatus, onVerifiedChange, pollIntervalMs]);

  async function handleResend() {
    if (!email) return;
    setIsSending(true);
    setError(false);
    try {
      await onResend(email);
      setResent(true);
    } catch {
      setError(true);
    } finally {
      setIsSending(false);
    }
  }

  const successMessage =
    verifiedStatus === 'already' ? labels.verifyEmailAlreadyVerified : labels.verifyEmailVerified;

  if (isVerified) {
    return (
      <div className={classNames.card}>
        <div className={classNames.success} role="status">
          {successMessage}
        </div>
        {email && (
          <p className="appkit-verify-email__email" style={{ fontWeight: 600, wordBreak: 'break-all', textAlign: 'center' }}>
            {email}
          </p>
        )}
        <p className={classNames.muted} style={{ textAlign: 'center' }}>
          {labels.verifyEmailReturnToSignIn}
        </p>
        <button
          type="button"
          className={classNames.submit}
          style={{ width: '100%', marginTop: '1rem' }}
          onClick={() => onGoToSignIn?.()}
        >
          {labels.verifyEmailGoToSignIn}
        </button>
      </div>
    );
  }

  return (
    <div className={classNames.card}>
      <p>{labels.verifyEmailInstructions}</p>
      {email && (
        <p className="appkit-verify-email__email" style={{ fontWeight: 600, wordBreak: 'break-all' }}>
          {email}
        </p>
      )}
      <p className={classNames.muted}>{labels.verifyEmailCheckSpam}</p>
      <p className={classNames.muted}>{labels.verifyEmailAnyBrowser}</p>
      {verifiedStatus === 'invalid' && (
        <p className={classNames.error} role="alert">
          {labels.verifyEmailInvalidLink}
        </p>
      )}
      {resent && (
        <div className={classNames.success} role="status">
          {labels.verifyEmailResent}
        </div>
      )}
      {error && (
        <p className={classNames.error} role="alert">
          {labels.genericError}
        </p>
      )}
      {email && (
        <button
          type="button"
          className={classNames.oauthButton ?? classNames.submit}
          style={{ width: '100%', marginTop: '1rem' }}
          onClick={handleResend}
          disabled={isSending}
          aria-busy={isSending}
        >
          {isSending ? <LoadingButtonContent label={labels.sending} /> : labels.verifyEmailResend}
        </button>
      )}
      <p className="appkit-login-workflow__back-row">
        {renderBackToSignIn?.(classNames.link) ?? null}
      </p>
    </div>
  );
}
