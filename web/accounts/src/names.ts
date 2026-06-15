import type { AccountUser } from './types';

export function accountUserFullName(user: Pick<AccountUser, 'firstName' | 'lastName'>): string {
  return [user.firstName?.trim(), user.lastName?.trim()].filter(Boolean).join(' ');
}

/** Derive 1–2 letter initials from first/last name, then email. */
export function accountUserInitials(user: Pick<AccountUser, 'firstName' | 'lastName' | 'email'>): string {
  const first = user.firstName?.trim();
  const last = user.lastName?.trim();
  if (first && last) {
    return `${first[0] ?? ''}${last[0] ?? ''}`.toUpperCase();
  }

  const name = accountUserFullName(user);
  if (name) {
    const parts = name.split(/\s+/).filter(Boolean);
    if (parts.length >= 2) {
      return `${parts[0]![0] ?? ''}${parts[parts.length - 1]![0] ?? ''}`.toUpperCase();
    }
    if (parts[0]!.length >= 2) {
      return parts[0]!.slice(0, 2).toUpperCase();
    }
    return parts[0]!.slice(0, 1).toUpperCase();
  }

  const local = user.email.split('@')[0] ?? '';
  if (local.length >= 2) {
    return local.slice(0, 2).toUpperCase();
  }
  return (local[0] ?? '?').toUpperCase();
}
