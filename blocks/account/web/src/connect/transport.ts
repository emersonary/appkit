import { createClient, type Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import type { AccountsSessionStorage } from './storage';

export type CreateConnectTransportOptions = {
  baseUrl?: string;
  storage: AccountsSessionStorage;
};

export function resolveAccountsApiBaseUrl(configured?: string): string {
  const trimmed = (configured ?? '').replace(/\/$/, '');
  return trimmed;
}

export function isAccountsApiConfigured(configured?: string): boolean {
  const base = resolveAccountsApiBaseUrl(configured);
  return base === '' || Boolean(base);
}

export function accountsHttpUrl(path: string, configured?: string): string {
  const base = resolveAccountsApiBaseUrl(configured);
  const normalized = path.startsWith('/') ? path : `/${path}`;
  return base ? `${base}${normalized}` : normalized;
}

export function createAccountsConnectTransport({ baseUrl = '', storage }: CreateConnectTransportOptions) {
  const accountSessionInterceptor: Interceptor = (next) => async (req) => {
    const session = storage.read();
    if (session?.accessToken) {
      req.header.set('Authorization', `Bearer ${session.accessToken}`);
    }
    return next(req);
  };

  return createConnectTransport({
    baseUrl,
    interceptors: [accountSessionInterceptor],
  });
}

export function createAccountsClient<T extends Parameters<typeof createClient>[0]>(
  service: T,
  options: CreateConnectTransportOptions,
) {
  return createClient(service, createAccountsConnectTransport(options));
}
