import { useState, type FormEvent, type ReactNode } from 'react';
import { useAccountsConfig } from './context';
import { FieldLabel } from './FieldLabel';
import { FieldIcon, RegisterTabIcon, SignInTabIcon, SubmitLoginIcon, SubmitRegisterIcon } from './fieldIcons';
import { LoadingSpinner } from './LoadingSpinner';
import { OAuthButtonGroup } from './OAuthButtonGroup';
import type { AccountsClassNames, AccountsUILabels } from './types';
import './loginWorkflow.css';

export type SignInUpMode = 'login' | 'register';

export type SignInUpPanelProps = {
  mode: SignInUpMode;
  onModeChange: (mode: SignInUpMode) => void;
  isLoading?: boolean;
  errorMessage?: ReactNode;
  verifiedBanner?: ReactNode;
  onLogin: (email: string, password: string) => Promise<void>;
  onRegister: (email: string, password: string, firstName: string, lastName?: string) => Promise<void>;
  onGoogleClick?: () => void;
  showGoogle?: boolean;
  renderForgotPasswordLink?: (className?: string) => ReactNode;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
  googleOAuthUrl?: string;
};

export function SignInUpPanel({
  mode,
  onModeChange,
  isLoading = false,
  errorMessage,
  verifiedBanner,
  onLogin,
  onRegister,
  onGoogleClick,
  showGoogle = true,
  renderForgotPasswordLink,
  labels: labelsProp,
  classNames: classNamesProp,
  googleOAuthUrl: googleOAuthUrlProp,
}: SignInUpPanelProps) {
  const ctx = useAccountsConfig();
  const labels = { ...ctx.labels, ...labelsProp };
  const classNames = { ...ctx.classNames, ...classNamesProp };
  const googleOAuthUrl = googleOAuthUrlProp ?? ctx.googleOAuthUrl;

  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);
  const [blockAutofill, setBlockAutofill] = useState(true);

  function switchMode(next: SignInUpMode) {
    onModeChange(next);
    setFirstName('');
    setLastName('');
    setEmail('');
    setPassword('');
    setConfirmPassword('');
    setLocalError(null);
    setBlockAutofill(true);
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLocalError(null);

    if (mode === 'register' && password !== confirmPassword) {
      setLocalError(labels.passwordMismatch ?? null);
      return;
    }

    if (mode === 'login') {
      await onLogin(email, password);
    } else {
      await onRegister(email, password, firstName, lastName || undefined);
    }
  }

  const showOAuth = mode === 'login' && showGoogle && Boolean(googleOAuthUrl || onGoogleClick);
  const requiredMarker = labels.requiredMarker;

  return (
    <>
      <div className="appkit-login-workflow__tabs">
        <button
          type="button"
          className={`appkit-login-workflow__tab${mode === 'login' ? ' active' : ''}`}
          onClick={() => switchMode('login')}
        >
          <SignInTabIcon />
          {labels.signIn ?? labels.loginTitle}
        </button>
        <button
          type="button"
          className={`appkit-login-workflow__tab${mode === 'register' ? ' active' : ''}`}
          onClick={() => switchMode('register')}
        >
          <RegisterTabIcon />
          {labels.registerTitle}
        </button>
      </div>
      {(errorMessage || localError) && (
        <p className={classNames.error} role="alert">
          {errorMessage ?? localError}
        </p>
      )}
      {verifiedBanner}
      <form className={classNames.card ?? classNames.form} onSubmit={handleSubmit} autoComplete="off" key={mode}>
        <div className={classNames.field}>
          <FieldLabel
            className={classNames.label}
            htmlFor="accounts-workflow-email"
            required
            requiredMarker={requiredMarker}
            icon={<FieldIcon kind="email" />}
          >
            {labels.email}
          </FieldLabel>
          <input
            id="accounts-workflow-email"
            name="accounts-workflow-email"
            className={classNames.input}
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onFocus={() => setBlockAutofill(false)}
            readOnly={blockAutofill}
            required
            autoComplete="off"
          />
        </div>
        {mode === 'register' && (
          <>
            <div className={classNames.field}>
              <FieldLabel
                className={classNames.label}
                htmlFor="accounts-workflow-first-name"
                required
                requiredMarker={requiredMarker}
                icon={<FieldIcon kind="firstName" />}
              >
                {labels.firstName}
              </FieldLabel>
              <input
                id="accounts-workflow-first-name"
                name="accounts-workflow-first-name"
                className={classNames.input}
                type="text"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                onFocus={() => setBlockAutofill(false)}
                readOnly={blockAutofill}
                required
                autoComplete="given-name"
              />
            </div>
            <div className={classNames.field}>
              <FieldLabel
                className={classNames.label}
                htmlFor="accounts-workflow-last-name"
                icon={<FieldIcon kind="lastName" />}
              >
                {labels.lastName}
              </FieldLabel>
              <input
                id="accounts-workflow-last-name"
                name="accounts-workflow-last-name"
                className={classNames.input}
                type="text"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                onFocus={() => setBlockAutofill(false)}
                readOnly={blockAutofill}
                autoComplete="family-name"
              />
            </div>
          </>
        )}
        <div className={classNames.field}>
          <FieldLabel
            className={classNames.label}
            htmlFor="accounts-workflow-password"
            required
            requiredMarker={requiredMarker}
            icon={<FieldIcon kind="password" />}
          >
            {labels.password}
          </FieldLabel>
          <input
            id="accounts-workflow-password"
            name="accounts-workflow-password"
            className={classNames.input}
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            onFocus={() => setBlockAutofill(false)}
            readOnly={blockAutofill}
            required
            minLength={6}
            autoComplete="new-password"
          />
          {mode === 'login' && renderForgotPasswordLink && (
            <div className={classNames.forgotRow ?? 'appkit-login-workflow__forgot-row'}>
              {renderForgotPasswordLink(classNames.link)}
            </div>
          )}
        </div>
        {mode === 'register' && (
          <div className={classNames.field}>
            <FieldLabel
              className={classNames.label}
              htmlFor="accounts-workflow-confirm-password"
              required
              requiredMarker={requiredMarker}
              icon={<FieldIcon kind="confirmPassword" />}
            >
              {labels.confirmPassword}
            </FieldLabel>
            <input
              id="accounts-workflow-confirm-password"
              name="accounts-workflow-confirm-password"
              className={classNames.input}
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              onFocus={() => setBlockAutofill(false)}
              readOnly={blockAutofill}
              required
              minLength={6}
              autoComplete="new-password"
            />
          </div>
        )}
        <button
          className={[classNames.submit, 'appkit-login-workflow__submit'].filter(Boolean).join(' ')}
          type="submit"
          disabled={isLoading}
          aria-busy={isLoading}
        >
          {isLoading ? (
            <LoadingSpinner label={labels.signingIn} size="sm" />
          ) : mode === 'login' ? (
            <>
              <SubmitLoginIcon />
              <span>{labels.logInAction ?? labels.loginSubmit}</span>
            </>
          ) : (
            <>
              <SubmitRegisterIcon />
              <span>{labels.createAccount ?? labels.registerSubmit}</span>
            </>
          )}
        </button>
      </form>
      {showOAuth && (
        <>
          <p className={classNames.divider ?? 'appkit-login-workflow__divider'}>{labels.orDivider}</p>
          <OAuthButtonGroup
            onGoogleClick={onGoogleClick}
            googleOAuthUrl={googleOAuthUrl}
            labels={labels}
            classNames={{
              ...classNames,
              oauthButton: [classNames.oauthButton, 'appkit-login-workflow__oauth-btn'].filter(Boolean).join(' '),
            }}
          />
        </>
      )}
    </>
  );
}
