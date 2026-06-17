import type { ReactNode } from 'react';
import { useAccountsSession } from './AccountsProvider';

export function SessionGate({
  children,
  fallback = null,
}: {
  children: ReactNode;
  fallback?: ReactNode;
}) {
  const { hasAccount, isInitialized } = useAccountsSession();
  if (!isInitialized) return null;
  return hasAccount ? children : fallback;
}

export function GuestGate({
  children,
  fallback = null,
}: {
  children: ReactNode;
  fallback?: ReactNode;
}) {
  const { hasAccount, isInitialized } = useAccountsSession();
  if (!isInitialized) return null;
  return !hasAccount ? children : fallback;
}
