import type { ReactNode } from 'react';
import { useAccountsSession } from './AccountsProvider';

export function SessionGate({
  children,
  fallback = null,
}: {
  children: ReactNode;
  fallback?: ReactNode;
}) {
  const { isAuthenticated, isInitialized } = useAccountsSession();
  if (!isInitialized) return null;
  return isAuthenticated ? children : fallback;
}

export function GuestGate({
  children,
  fallback = null,
}: {
  children: ReactNode;
  fallback?: ReactNode;
}) {
  const { isAuthenticated, isInitialized } = useAccountsSession();
  if (!isInitialized) return null;
  return !isAuthenticated ? children : fallback;
}
