import type { AccountSession } from '../types';

export type AccountsSessionStorage = {
  read: () => AccountSession | null;
  write: (session: AccountSession | null) => void;
};

const DEFAULT_KEY = 'appkit_auth_session';

export function createLocalStorageSessionStorage(key = DEFAULT_KEY): AccountsSessionStorage {
  return {
    read() {
      try {
        const raw = localStorage.getItem(key);
        if (!raw) return null;
        return JSON.parse(raw) as AccountSession;
      } catch {
        return null;
      }
    },
    write(session) {
      if (session) {
        localStorage.setItem(key, JSON.stringify(session));
      } else {
        localStorage.removeItem(key);
      }
    },
  };
}
