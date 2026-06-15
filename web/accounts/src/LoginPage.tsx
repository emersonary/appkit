import { useState, type FormEvent, type ReactNode } from 'react';
import { useAccountsConfig } from './context';
import { OAuthButtonGroup } from './OAuthButtonGroup';
import type { AccountsClassNames, AccountsUILabels } from './types';

export type LoginPageProps = {
  initialMode?: 'login' | 'register';
  isLoading?: boolean;
  errorMessage?: ReactNode;
  verifiedBanner?: ReactNode;
  onLogin: (email: string, password: string) => Promise<void>;
  onRegister: (email: string, password: string) => Promise<void>;
  onGoogleClick?: () => void;
  forgotPasswordHref?: string;
  renderForgotPasswordLink?: (className?: string) => ReactNode;
  showRegister?: boolean;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
  googleOAuthUrl?: string;
};

export function LoginPage({
  initialMode = 'login',
  isLoading = false,
  errorMessage,
  verifiedBanner,
  onLogin,
  onRegister,
  onGoogleClick,
  forgotPasswordHref,
  renderForgotPasswordLink,
  showRegister = true,
  labels: labelsProp,
  classNames: classNamesProp,
  googleOAuthUrl: googleOAuthUrlProp,
}: LoginPageProps) {
  const ctx = useAccountsConfig();
  const labels = { ...ctx.labels, ...labelsProp };
  const classNames = { ...ctx.classNames, ...classNamesProp };
  const googleOAuthUrl = googleOAuthUrlProp ?? ctx.googleOAuthUrl;
  const [mode, setMode] = useState<'login' | 'register'>(initialMode);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLocalError(null);
    if (mode === 'register' && password !== confirmPassword) {
      setLocalError(labels.passwordMismatch);
      return;
    }
    if (mode === 'login') {
      await onLogin(email, password);
    } else {
      await onRegister(email, password);
    }
  }

  return (
    <div className={classNames.page}>
      <div className={classNames.card}>
        <h1 className={classNames.title}>{mode === 'login' ? labels.loginTitle : labels.registerTitle}</h1>
        {verifiedBanner}
        {(errorMessage || localError) && (
          <p className={classNames.error} role="alert">
            {errorMessage ?? localError}
          </p>
        )}
        <form className={classNames.form} onSubmit={handleSubmit}>
          <div className={classNames.field}>
            <label className={classNames.label} htmlFor="accounts-email">
              {labels.email}
            </label>
            <input
              id="accounts-email"
              className={classNames.input}
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className={classNames.field}>
            <label className={classNames.label} htmlFor="accounts-password">
              {labels.password}
            </label>
            <input
              id="accounts-password"
              className={classNames.input}
              type="password"
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
            />
          </div>
          {mode === 'register' && (
            <div className={classNames.field}>
              <label className={classNames.label} htmlFor="accounts-confirm-password">
                {labels.confirmPassword}
              </label>
              <input
                id="accounts-confirm-password"
                className={classNames.input}
                type="password"
                autoComplete="new-password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                minLength={6}
              />
            </div>
          )}
          {mode === 'login' && (
            <div className={classNames.forgotRow}>
              {renderForgotPasswordLink?.(classNames.link) ??
                (forgotPasswordHref ? (
                  <a className={classNames.link} href={forgotPasswordHref}>
                    {labels.forgotPassword}
                  </a>
                ) : null)}
            </div>
          )}
          <button className={classNames.submit} type="submit" disabled={isLoading} style={{ width: '100%' }}>
            {mode === 'login' ? labels.loginSubmit : labels.registerSubmit}
          </button>
        </form>
        {googleOAuthUrl && (
          <>
            <p className={classNames.divider}>{labels.orDivider}</p>
            <OAuthButtonGroup
              onGoogleClick={onGoogleClick}
              googleOAuthUrl={googleOAuthUrl}
              labels={labels}
              classNames={classNames}
            />
          </>
        )}
        {showRegister && (
          <div className={classNames.modeSwitch}>
            <button
              type="button"
              className={classNames.link}
              onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
            >
              {mode === 'login' ? labels.switchToRegister : labels.switchToLogin}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
