import type { Account } from './types';

export function accountFullName(account: Pick<Account, 'firstName' | 'lastName'>): string {
  return [account.firstName?.trim(), account.lastName?.trim()].filter(Boolean).join(' ');
}

/** Initials for avatar fallback (up to two letters). */
export function accountInitials(account: Pick<Account, 'firstName' | 'lastName' | 'email'>): string {
  const first = account.firstName?.trim();
  const last = account.lastName?.trim();
  if (first && last) {
    return `${first[0] ?? ''}${last[0] ?? ''}`.toUpperCase();
  }

  const name = accountFullName(account);
  if (name) {
    const parts = name.split(/\s+/).filter(Boolean);
    if (parts.length >= 2) {
      return `${parts[0]![0] ?? ''}${parts[1]![0] ?? ''}`.toUpperCase();
    }
    if (parts[0]) {
      return parts[0].slice(0, 2).toUpperCase();
    }
  }

  const local = account.email.split('@')[0] ?? '';
  return local.slice(0, 2).toUpperCase() || '?';
}
