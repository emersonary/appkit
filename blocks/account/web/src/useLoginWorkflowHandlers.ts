import { useCallback } from 'react';
import { useAccountsSession } from './AccountsProvider';
import { useAccountsConfig } from './context';
import { AccountsError, AccountsErrorCode } from './connect/errors';
import { googleAccountsLoginUrl, isAccountsApiConfigured } from './connect';

export type LoginWorkflowHandlers = {
  isLoading: boolean;
  onLogin: (email: string, password: string) => Promise<void>;
  onRegister: (email: string, password: string, firstName: string, lastName?: string) => Promise<{
    verificationRequired: boolean;
    email: string;
  }>;
  onRequestReset: (email: string) => Promise<void>;
  onResetPassword: (token: string, password: string) => Promise<void>;
  onFetchVerificationStatus: (email: string) => Promise<boolean>;
  onResendVerificationEmail: (email: string) => Promise<void>;
  onGoogleClick?: () => void;
  showGoogle: boolean;
  classifyResetError: (err: unknown) => string | null;
};

export type UseLoginWorkflowHandlersOptions = {
  apiBaseUrl?: string;
  onAccountError?: (err: unknown) => void;
  mapResetError?: (err: unknown) => string | null;
};

function defaultClassifyResetError(err: unknown): string | null {
  if (err instanceof AccountsError && err.code === AccountsErrorCode.INVALID_ARGUMENT) {
    return 'invalid-link';
  }
  return 'generic';
}

function browserLanguage(): string | undefined {
  if (typeof navigator === 'undefined') {
    return undefined;
  }
  const code = navigator.language.split('-')[0]?.toLowerCase();
  return code || undefined;
}

export function useLoginWorkflowHandlers({
  apiBaseUrl = '',
  onAccountError,
  mapResetError,
}: UseLoginWorkflowHandlersOptions = {}): LoginWorkflowHandlers {
  const {
    isLoading,
    login,
    register,
    requestPasswordReset,
    resetPassword,
    getVerificationStatus,
    resendVerificationEmail,
  } = useAccountsSession();
  const { oauth, registrationEnabled } = useAccountsConfig();

  const wrap = useCallback(
    async <T,>(fn: () => Promise<T>): Promise<T> => {
      try {
        return await fn();
      } catch (err) {
        onAccountError?.(err);
        throw err;
      }
    },
    [onAccountError],
  );

  const onLogin = useCallback(
    (email: string, password: string) => wrap(() => login(email, password)),
    [login, wrap],
  );

  const onRegister = useCallback(
    (email: string, password: string, firstName: string, lastName?: string) =>
      wrap(() => register(email, password, firstName, lastName, browserLanguage())),
    [register, wrap],
  );

  const onRequestReset = useCallback(
    (email: string) => wrap(() => requestPasswordReset(email)),
    [requestPasswordReset, wrap],
  );

  const onResetPassword = useCallback(
    (token: string, password: string) => wrap(() => resetPassword(token, password)),
    [resetPassword, wrap],
  );

  const onFetchVerificationStatus = useCallback(
    (email: string) => getVerificationStatus(email),
    [getVerificationStatus],
  );

  const onResendVerificationEmail = useCallback(
    (email: string) => wrap(() => resendVerificationEmail(email)),
    [resendVerificationEmail, wrap],
  );

  const googleConfigured = isAccountsApiConfigured(apiBaseUrl);
  const oauthEnabled = oauth.enabled !== false;
  const googleOAuthEnabled = oauth.google !== false;
  const showGoogle =
    registrationEnabled && oauthEnabled && googleOAuthEnabled && googleConfigured;

  const onGoogleClick = showGoogle
    ? () => {
        window.location.href = googleAccountsLoginUrl(apiBaseUrl);
      }
    : undefined;

  const classifyResetError = useCallback(
    (err: unknown) => mapResetError?.(err) ?? defaultClassifyResetError(err),
    [mapResetError],
  );

  return {
    isLoading,
    onLogin,
    onRegister,
    onRequestReset,
    onResetPassword,
    onFetchVerificationStatus,
    onResendVerificationEmail,
    onGoogleClick,
    showGoogle,
    classifyResetError,
  };
}
