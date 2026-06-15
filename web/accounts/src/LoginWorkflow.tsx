import type { ReactNode } from 'react';
import { ForgotPasswordPage } from './ForgotPasswordPage';
import { ResetPasswordPage } from './ResetPasswordPage';
import { SignInUpPanel, type SignInUpMode } from './SignInUpPanel';
import { VerifyEmailPage } from './VerifyEmailPage';
import type { AccountsClassNames, AccountsUILabels } from './types';
import './loginWorkflow.css';

export type LoginWorkflowStep = 'sign-in' | 'sign-up' | 'forgot' | 'reset' | 'verify-email';

export type LoginWorkflowProps = {
  step: LoginWorkflowStep;
  onStepChange?: (step: LoginWorkflowStep) => void;
  resetToken?: string;
  isLoading?: boolean;
  errorMessage?: ReactNode;
  verifiedBanner?: ReactNode;
  onLogin: (email: string, password: string) => Promise<void>;
  onRegister: (email: string, password: string, firstName: string, lastName?: string) => Promise<void>;
  onRequestReset: (email: string) => Promise<void>;
  onResetPassword: (token: string, password: string) => Promise<void>;
  classifyResetError?: (err: unknown) => string | null;
  onResetSuccess?: () => void;
  onGoToSignIn?: () => void;
  onGoogleClick?: () => void;
  showGoogle?: boolean;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
  googleOAuthUrl?: string;
  renderForgotPasswordLink?: (className?: string) => ReactNode;
  renderBackToSignIn?: (className?: string) => ReactNode;
  renderRequestNewReset?: (className?: string) => ReactNode;
  verifyEmail?: string;
  verifyEmailStatus?: string | null;
  onFetchVerificationStatus?: (email: string) => Promise<boolean>;
  onResendVerificationEmail?: (email: string) => Promise<void>;
  onVerifiedChange?: (verified: boolean) => void;
};

function stepToMode(step: LoginWorkflowStep): SignInUpMode {
  return step === 'sign-up' ? 'register' : 'login';
}

function modeToStep(mode: SignInUpMode): LoginWorkflowStep {
  return mode === 'register' ? 'sign-up' : 'sign-in';
}

export function LoginWorkflow({
  step,
  onStepChange,
  resetToken = '',
  isLoading = false,
  errorMessage,
  verifiedBanner,
  onLogin,
  onRegister,
  onRequestReset,
  onResetPassword,
  classifyResetError,
  onResetSuccess,
  onGoToSignIn,
  onGoogleClick,
  showGoogle = true,
  labels,
  classNames,
  googleOAuthUrl,
  renderForgotPasswordLink,
  renderBackToSignIn,
  renderRequestNewReset,
  verifyEmail = '',
  verifyEmailStatus = null,
  onFetchVerificationStatus,
  onResendVerificationEmail,
  onVerifiedChange,
}: LoginWorkflowProps) {
  if (step === 'verify-email' && onFetchVerificationStatus && onResendVerificationEmail) {
    return (
      <VerifyEmailPage
        email={verifyEmail}
        verifiedStatus={verifyEmailStatus}
        onFetchStatus={onFetchVerificationStatus}
        onResend={onResendVerificationEmail}
        onGoToSignIn={onGoToSignIn}
        onVerifiedChange={onVerifiedChange}
        labels={labels}
        classNames={classNames}
        renderBackToSignIn={renderBackToSignIn}
      />
    );
  }

  if (step === 'forgot') {
    return (
      <ForgotPasswordPage
        onRequestReset={onRequestReset}
        labels={labels}
        classNames={classNames}
        renderBackToLogin={renderBackToSignIn}
      />
    );
  }

  if (step === 'reset') {
    return (
      <ResetPasswordPage
        token={resetToken}
        onResetPassword={onResetPassword}
        classifyError={classifyResetError}
        onSuccess={onResetSuccess}
        onGoToSignIn={onGoToSignIn}
        labels={labels}
        classNames={classNames}
        renderBackToLogin={renderBackToSignIn}
        renderRequestNew={renderRequestNewReset}
      />
    );
  }

  return (
    <SignInUpPanel
      mode={stepToMode(step)}
      onModeChange={(mode) => onStepChange?.(modeToStep(mode))}
      isLoading={isLoading}
      errorMessage={errorMessage}
      verifiedBanner={verifiedBanner}
      onLogin={onLogin}
      onRegister={onRegister}
      onGoogleClick={onGoogleClick}
      showGoogle={showGoogle}
      renderForgotPasswordLink={renderForgotPasswordLink}
      labels={labels}
      classNames={classNames}
      googleOAuthUrl={googleOAuthUrl}
    />
  );
}
