import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { AccountsConfigContext, defaultLabels } from './context';
import type { AccountSession, AccountUser, AccountsProviderProps } from './types';

export type AccountsSessionContextValue = {
  user: AccountUser | null;
  session: AccountSession | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isInitialized: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (
    email: string,
    password: string,
    firstName: string,
    lastName?: string,
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
  authClient,
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
      googleOAuthUrl,
      settingsHref,
      onSettingsClick,
    }),
    [tenancy, labels, classNames, googleOAuthUrl, settingsHref, onSettingsClick],
  );

  const refreshSession = useCallback(async () => {
    const next = await authClient.getSession();
    setSession(next);
  }, [authClient]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const existing = await authClient.getSession();
        if (!cancelled) setSession(existing);
      } finally {
        if (!cancelled) setIsInitialized(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [authClient]);

  const login = useCallback(
    async (email: string, password: string) => {
      setIsLoading(true);
      try {
        const next = await authClient.login(email, password);
        setSession(next);
      } finally {
        setIsLoading(false);
      }
    },
    [authClient],
  );

  const register = useCallback(
    async (email: string, password: string, firstName: string, lastName?: string) => {
      setIsLoading(true);
      try {
        const result = await authClient.register(email, password, firstName, lastName);
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
    [authClient, refreshSession],
  );

  const logout = useCallback(async () => {
    setIsLoading(true);
    try {
      await authClient.logout();
      setSession(null);
    } finally {
      setIsLoading(false);
    }
  }, [authClient]);

  const completeOAuthLogin = useCallback(
    async (accessToken: string) => {
      if (!authClient.completeOAuthLogin) {
        throw new Error('completeOAuthLogin is not configured');
      }
      setIsLoading(true);
      try {
        const next = await authClient.completeOAuthLogin(accessToken);
        setSession(next);
      } finally {
        setIsLoading(false);
      }
    },
    [authClient],
  );

  const requestPasswordReset = useCallback(
    async (email: string) => {
      if (!authClient.requestPasswordReset) {
        throw new Error('requestPasswordReset is not configured');
      }
      setIsLoading(true);
      try {
        await authClient.requestPasswordReset(email);
      } finally {
        setIsLoading(false);
      }
    },
    [authClient],
  );

  const resetPassword = useCallback(
    async (token: string, newPassword: string) => {
      if (!authClient.resetPassword) {
        throw new Error('resetPassword is not configured');
      }
      setIsLoading(true);
      try {
        await authClient.resetPassword(token, newPassword);
      } finally {
        setIsLoading(false);
      }
    },
    [authClient],
  );

  const getVerificationStatus = useCallback(
    async (email: string) => {
      if (!authClient.getVerificationStatus) {
        throw new Error('getVerificationStatus is not configured');
      }
      return authClient.getVerificationStatus(email);
    },
    [authClient],
  );

  const resendVerificationEmail = useCallback(
    async (email: string) => {
      if (!authClient.resendVerificationEmail) {
        throw new Error('resendVerificationEmail is not configured');
      }
      setIsLoading(true);
      try {
        await authClient.resendVerificationEmail(email);
      } finally {
        setIsLoading(false);
      }
    },
    [authClient],
  );

  const sessionValue = useMemo<AccountsSessionContextValue>(
    () => ({
      user: session?.user ?? null,
      session,
      isAuthenticated: !!session,
      isLoading,
      isInitialized,
      login,
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

export function useAccountUser() {
  return useAccountsSession().user;
}
