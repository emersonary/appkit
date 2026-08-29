import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { AccountsConfigContext, defaultLabels } from './context';
import type { AccountSession, Account, AccountsProviderProps } from './types';
import { defaultAccountsOAuthConfig } from './types';

export type AccountsSessionContextValue = {
  account: Account | null;
  session: AccountSession | null;
  hasAccount: boolean;
  isLoading: boolean;
  isInitialized: boolean;
  login: (email: string, password: string) => Promise<void>;
  adminLogin: (email: string, password: string) => Promise<void>;
  register: (
    email: string,
    password: string,
    firstName: string,
    lastName?: string,
    language?: string,
  ) => Promise<{ verificationRequired: boolean; email: string }>;
  logout: () => Promise<void>;
  completeOAuthLogin: (accessToken: string) => Promise<void>;
  refreshSession: () => Promise<void>;
  requestPasswordReset: (email: string) => Promise<void>;
  resetPassword: (token: string, newPassword: string) => Promise<void>;
  getVerificationStatus: (email: string) => Promise<boolean>;
  resendVerificationEmail: (email: string) => Promise<void>;
};

const AccountsSessionCtx = createContext<AccountsSessionContextValue | null>(null);

export { AccountsSessionCtx };

export function AccountsProvider({
  children,
  tenancy,
  accountClient,
  oauth,
  registrationEnabled = true,
  googleOAuthUrl,
  labels,
  classNames,
  settingsHref,
  onSettingsClick,
}: AccountsProviderProps) {
  const [session, setSession] = useState<AccountSession | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);

  const configValue = useMemo(
    () => ({
      tenancy,
      labels: { ...defaultLabels, ...labels },
      classNames: classNames ?? {},
      oauth: { ...defaultAccountsOAuthConfig, ...oauth },
      registrationEnabled,
      googleOAuthUrl,
      settingsHref,
      onSettingsClick,
    }),
    [tenancy, labels, classNames, oauth, registrationEnabled, googleOAuthUrl, settingsHref, onSettingsClick],
  );

  const refreshSession = useCallback(async () => {
    const next = await accountClient.getSession();
    setSession(next);
  }, [accountClient]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const existing = await accountClient.getSession();
        if (!cancelled) setSession(existing);
      } finally {
        if (!cancelled) setIsInitialized(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [accountClient]);

  const login = useCallback(
    async (email: string, password: string) => {
      setIsLoading(true);
      try {
        const next = await accountClient.login(email, password);
        setSession(next);
      } finally {
        setIsLoading(false);
      }
    },
    [accountClient],
  );

  const adminLogin = useCallback(
    async (email: string, password: string) => {
      setIsLoading(true);
      try {
        const next = await accountClient.adminLogin(email, password);
        setSession(next);
      } finally {
        setIsLoading(false);
      }
    },
    [accountClient],
  );

  const register = useCallback(
    async (email: string, password: string, firstName: string, lastName?: string, language?: string) => {
      setIsLoading(true);
      try {
        const result = await accountClient.register(email, password, firstName, lastName, language);
        if (!result.verificationRequired) {
          await refreshSession();
        } else {
          setSession(null);
        }
        return result;
      } finally {
        setIsLoading(false);
      }
    },
    [accountClient, refreshSession],
  );

  const logout = useCallback(async () => {
    setIsLoading(true);
    try {
      await accountClient.logout();
      setSession(null);
    } finally {
      setIsLoading(false);
    }
  }, [accountClient]);

  const completeOAuthLogin = useCallback(
    async (accessToken: string) => {
      if (!accountClient.completeOAuthLogin) {
        throw new Error('completeOAuthLogin is not configured');
      }
      setIsLoading(true);
      try {
        const next = await accountClient.completeOAuthLogin(accessToken);
        setSession(next);
      } finally {
        setIsLoading(false);
      }
    },
    [accountClient],
  );

  const requestPasswordReset = useCallback(
    async (email: string) => {
      if (!accountClient.requestPasswordReset) {
        throw new Error('requestPasswordReset is not configured');
      }
      setIsLoading(true);
      try {
        await accountClient.requestPasswordReset(email);
      } finally {
        setIsLoading(false);
      }
    },
    [accountClient],
  );

  const resetPassword = useCallback(
    async (token: string, newPassword: string) => {
      if (!accountClient.resetPassword) {
        throw new Error('resetPassword is not configured');
      }
      setIsLoading(true);
      try {
        await accountClient.resetPassword(token, newPassword);
      } finally {
        setIsLoading(false);
      }
    },
    [accountClient],
  );

  const getVerificationStatus = useCallback(
    async (email: string) => {
      if (!accountClient.getVerificationStatus) {
        throw new Error('getVerificationStatus is not configured');
      }
      return accountClient.getVerificationStatus(email);
    },
    [accountClient],
  );

  const resendVerificationEmail = useCallback(
    async (email: string) => {
      if (!accountClient.resendVerificationEmail) {
        throw new Error('resendVerificationEmail is not configured');
      }
      setIsLoading(true);
      try {
        await accountClient.resendVerificationEmail(email);
      } finally {
        setIsLoading(false);
      }
    },
    [accountClient],
  );

  const sessionValue = useMemo<AccountsSessionContextValue>(
    () => ({
      account: session?.account ?? null,
      session,
      hasAccount: !!session,
      isLoading,
      isInitialized,
      login,
      adminLogin,
      register,
      logout,
      completeOAuthLogin,
      refreshSession,
      requestPasswordReset,
      resetPassword,
      getVerificationStatus,
      resendVerificationEmail,
    }),
    [
      session,
      isLoading,
      isInitialized,
      login,
      adminLogin,
      register,
      logout,
      completeOAuthLogin,
      refreshSession,
      requestPasswordReset,
      resetPassword,
      getVerificationStatus,
      resendVerificationEmail,
    ],
  );

  return (
    <AccountsConfigContext.Provider value={configValue}>
      <AccountsSessionCtx.Provider value={sessionValue}>{children}</AccountsSessionCtx.Provider>
    </AccountsConfigContext.Provider>
  );
}

export function useAccountsSession() {
  const ctx = useContext(AccountsSessionCtx);
  if (!ctx) {
    throw new Error('useAccountsSession must be used within AccountsProvider');
  }
  return ctx;
}

export function useSessionAccount() {
  return useAccountsSession().account;
}
