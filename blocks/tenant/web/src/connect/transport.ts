import { createClient, type Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import type { TenantsSessionStorage } from '../types';

export type CreateTenantsTransportOptions = {
  baseUrl?: string;
  storage: TenantsSessionStorage;
};

export function createTenantsConnectTransport({ baseUrl = '', storage }: CreateTenantsTransportOptions) {
  const authInterceptor: Interceptor = (next) => async (req) => {
    const token = storage.readAccessToken();
    if (token) {
      req.header.set('Authorization', `Bearer ${token}`);
    }
    return next(req);
  };

  return createConnectTransport({
    baseUrl,
    interceptors: [authInterceptor],
  });
}

export function createTenantsClient<T extends Parameters<typeof createClient>[0]>(
  service: T,
  options: CreateTenantsTransportOptions,
) {
  return createClient(service, createTenantsConnectTransport(options));
}
