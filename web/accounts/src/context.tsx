import { createContext, useContext } from 'react';
import type { AccountsClassNames, AccountsUILabels, AccountsTenancyConfig } from './types';

export type AccountsContextValue = {
  tenancy: AccountsTenancyConfig;
  labels: Required<AccountsUILabels>;
  classNames: AccountsClassNames;
  googleOAuthUrl?: string;
  settingsHref?: string;
  onSettingsClick?: () => void;
};

const defaultLabels: Required<AccountsUILabels> = {
  signIn: 'Sign in',
  signOut: 'Sign out',
  accountMenu: 'Account',
  loginTitle: 'Sign in',
  registerTitle: 'Create account',
  email: 'Email',
  firstName: 'First name',
  lastName: 'Last name',
  password: 'Password',
  confirmPassword: 'Confirm password',
  loginSubmit: 'Sign in',
  registerSubmit: 'Create account',
  logInAction: 'Sign in',
  createAccount: 'Create account',
  signingIn: 'Please wait…',
  requiredMarker: ' (*)',
  verifyEmailInstructions: 'We sent a verification link to:',
  verifyEmailCheckSpam: 'If you do not see it, check your spam folder.',
  verifyEmailAnyBrowser: 'You can open the link in any browser. This page will update when your email is verified.',
  verifyEmailResend: 'Resend verification email',
  verifyEmailResent: 'Verification email sent.',
  verifyEmailBackToLogin: 'Back to sign in',
  verifyEmailVerified: 'Your email is verified. You can sign in now.',
  verifyEmailAlreadyVerified: 'This email was already verified. You can sign in now.',
  verifyEmailInvalidLink: 'That verification link is invalid or expired. Request a new one below.',
  verifyEmailReturnToSignIn: 'Return to the browser where you started, or sign in here.',
  verifyEmailGoToSignIn: 'Go to sign in',
  switchToRegister: 'Create an account',
  switchToLogin: 'Already have an account? Sign in',
  forgotPassword: 'Forgot password?',
  continueWithGoogle: 'Continue with Google',
  orDivider: 'or',
  passwordMismatch: 'Passwords do not match',
  forgotPasswordTitle: 'Reset your password',
  forgotPasswordInstructions: 'Enter your email and we will send you a reset link.',
  forgotPasswordSubmit: 'Send reset link',
  forgotPasswordSent: 'If an account exists for that email, we sent a reset link.',
  forgotPasswordCheckSpam: 'Check your spam folder if you do not see it.',
  backToLogin: 'Back to sign in',
  resetPasswordTitle: 'Choose a new password',
  resetPasswordSubmit: 'Update password',
  resetPasswordSuccess: 'Your password has been updated.',
  resetPasswordInvalidLink: 'This reset link is invalid or has expired.',
  resetPasswordRequestNew: 'Request a new link',
  goToSignIn: 'Go to sign in',
  sending: 'Please wait…',
  genericError: 'Something went wrong. Please try again.',
};

export const AccountsConfigContext = createContext<AccountsContextValue>({
  tenancy: { enabled: false },
  labels: defaultLabels,
  classNames: {},
});

export function useAccountsConfig() {
  return useContext(AccountsConfigContext);
}

export { defaultLabels };
