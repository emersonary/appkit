import { createContext, useContext, type ReactNode } from 'react';
import type { MessengerAdapter } from './adapters/MessengerAdapter';

const MessengerContext = createContext<MessengerAdapter | null>(null);

export type MessengerProviderProps = {
  adapter: MessengerAdapter;
  children: ReactNode;
};

export function MessengerProvider({ adapter, children }: MessengerProviderProps) {
  return <MessengerContext.Provider value={adapter}>{children}</MessengerContext.Provider>;
}

export function useMessengerAdapter(): MessengerAdapter {
  const adapter = useContext(MessengerContext);
  if (!adapter) {
    throw new Error('useMessengerAdapter must be used within MessengerProvider');
  }
  return adapter;
}
