export type Account = {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  avatarUrl?: string;
  isAdmin?: boolean;
};

export type AccountSession = {
  accessToken: string;
  account: Account;
  tenantId?: string;
};

export type AccountsTenancyConfig = {
  enabled: boolean;
};

export type AccountsUILabels = {
  signIn?: string;
  signOut?: string;
  accountMenu?: string;
  loginTitle?: string;
  registerTitle?: string;
  email?: string;
  firstName?: string;
  lastName?: string;
  password?: string;
  confirmPassword?: string;
  loginSubmit?: string;
  registerSubmit?: string;
  logInAction?: string;
  createAccount?: string;
  signingIn?: string;
  requiredMarker?: string;
  verifyEmailInstructions?: string;
  verifyEmailCheckSpam?: string;
  verifyEmailAnyBrowser?: string;
  verifyEmailResend?: string;
  verifyEmailResent?: string;
  verifyEmailBackToLogin?: string;
  verifyEmailVerified?: string;
  verifyEmailAlreadyVerified?: string;
  verifyEmailInvalidLink?: string;
  verifyEmailReturnToSignIn?: string;
  verifyEmailGoToSignIn?: string;
  switchToRegister?: string;
  switchToLogin?: string;
  forgotPassword?: string;
  continueWithGoogle?: string;
  orDivider?: string;
  passwordMismatch?: string;
  forgotPasswordTitle?: string;
  forgotPasswordInstructions?: string;
  forgotPasswordSubmit?: string;
  forgotPasswordSent?: string;
  forgotPasswordCheckSpam?: string;
  backToLogin?: string;
  resetPasswordTitle?: string;
  resetPasswordSubmit?: string;
  resetPasswordSuccess?: string;
  resetPasswordInvalidLink?: string;
  resetPasswordRequestNew?: string;
  goToSignIn?: string;
  sending?: string;
  genericError?: string;
};

export type AccountsClassNames = {
  page?: string;
  card?: string;
  title?: string;
  form?: string;
  field?: string;
  label?: string;
  input?: string;
  error?: string;
  success?: string;
  muted?: string;
  instructions?: string;
  submit?: string;
  link?: string;
  divider?: string;
  oauthGroup?: string;
  oauthButton?: string;
  forgotRow?: string;
  footer?: string;
  modeSwitch?: string;
  trigger?: string;
  signInTrigger?: string;
  signInIcon?: string;
  panel?: string;
  panelHeader?: string;
  panelEmail?: string;
  menuItem?: string;
  avatar?: string;
  avatarImage?: string;
  avatarFallback?: string;
};

export type RegisterResult = {
  verificationRequired: boolean;
  email: string;
};

export type AccountsClient = {
  login: (email: string, password: string) => Promise<AccountSession>;
  register: (
    email: string,
    password: string,
    firstName: string,
    lastName?: string,
  ) => Promise<RegisterResult>;
  logout: () => Promise<void>;
  getSession: () => Promise<AccountSession | null>;
  completeOAuthLogin?: (accessToken: string) => Promise<AccountSession>;
  requestPasswordReset?: (email: string) => Promise<void>;
  resetPassword?: (token: string, newPassword: string) => Promise<void>;
  getVerificationStatus?: (email: string) => Promise<boolean>;
  resendVerificationEmail?: (email: string) => Promise<void>;
};

export type AccountsProviderProps = {
  children: React.ReactNode;
  tenancy: AccountsTenancyConfig;
  accountClient: AccountsClient;
  googleOAuthUrl?: string;
  labels?: AccountsUILabels;
  classNames?: AccountsClassNames;
  settingsHref?: string;
  onSettingsClick?: () => void;
  renderSettings?: (close: () => void) => React.ReactNode;
  renderPanelExtra?: (account: Account, close: () => void) => React.ReactNode;
};

import type React from 'react';
