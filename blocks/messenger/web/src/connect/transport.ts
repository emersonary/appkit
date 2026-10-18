import { createClient, type Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import type { MessengerSessionStorage } from '../types';

export type CreateMessengerTransportOptions = {
  baseUrl?: string;
  storage: MessengerSessionStorage;
  agentApiKey?: string;
  agentUserId?: string;
};

export function createMessengerConnectTransport({
  baseUrl = '',
  storage,
  agentApiKey,
  agentUserId,
}: CreateMessengerTransportOptions) {
  const authInterceptor: Interceptor = (next) => async (req) => {
    const token = storage.readAccessToken();
    if (token) {
      req.header.set('Authorization', `Bearer ${token}`);
    }
    if (agentApiKey) {
      req.header.set('X-Messenger-Agent-Key', agentApiKey);
    }
    if (agentUserId) {
      req.header.set('X-Messenger-Agent-User-Id', agentUserId);
    }
    return next(req);
  };

  return createConnectTransport({
    baseUrl,
    interceptors: [authInterceptor],
  });
}

export function createMessengerClient<T extends Parameters<typeof createClient>[0]>(
  service: T,
  options: CreateMessengerTransportOptions,
) {
  return createClient(service, createMessengerConnectTransport(options));
}
