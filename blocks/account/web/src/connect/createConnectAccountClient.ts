import { GetSessionRequest, LogoutRequest } from '../gen/v1/account_pb';
import { AccountService } from '../gen/v1/account_connect';
import type { AccountSession, AccountsClient } from '../types';
import { fromConnectError } from './connectErrors';
import { createAccountsClient } from './transport';
import type { AccountsSessionStorage } from './storage';

export type CreateConnectAccountClientOptions = {
  baseUrl?: string;
  storage: AccountsSessionStorage;
};

function toSession(response: {
  accessToken: string;
  account?: {
    id: string;
    email: string;
    firstName?: string;
    lastName?: string;
    avatarUrl?: string;
    isAdmin?: boolean;
    language?: string;
  };
}): AccountSession {
  return {
    accessToken: response.accessToken,
    account: {
      id: response.account?.id ?? '',
      email: response.account?.email ?? '',
      firstName: response.account?.firstName || undefined,
      lastName: response.account?.lastName || undefined,
      avatarUrl: response.account?.avatarUrl || undefined,
      isAdmin: response.account?.isAdmin ?? false,
      language: response.account?.language || undefined,
    },
  };
}

function isLegacyMockToken(token: string): boolean {
  return token.startsWith('mock-');
}

export function createConnectAccountClient({
  baseUrl = '',
  storage,
}: CreateConnectAccountClientOptions): AccountsClient {
  const client = createAccountsClient(AccountService, { baseUrl, storage });

  return {
    async login(email, password) {
      try {
        const response = await client.login({
          email: email.trim(),
          password,
        });
        const session = toSession(response);
        storage.write(session);
        return session;
      } catch (err) {
        throw fromConnectError(err);
      }
    },

    async register(email, password, firstName, lastName, language) {
      try {
        const response = await client.register({
          email: email.trim(),
          password,
          firstName: firstName.trim(),
          lastName: lastName?.trim() ?? '',
          language: language?.trim() ?? '',
        });
        if (response.verificationRequired) {
          storage.write(null);
          return {
            verificationRequired: true,
            email: response.session?.account?.email ?? email.trim(),
          };
        }
        const session = toSession(response.session!);
        storage.write(session);
        return {
          verificationRequired: false,
          email: session.account.email,
        };
      } catch (err) {
        throw fromConnectError(err);
      }
    },

    async logout() {
      storage.write(null);
      try {
        await client.logout(new LogoutRequest());
      } catch {
        /* JWT sessions are stateless; local state is already cleared. */
      }
    },

    async getSession() {
      const stored = storage.read();
      if (!stored?.accessToken || isLegacyMockToken(stored.accessToken)) {
        storage.write(null);
        return null;
      }

      try {
        const response = await client.getSession(new GetSessionRequest());
        const session = toSession(response);
        storage.write(session);
        return session;
      } catch {
        storage.write(null);
        return null;
      }
    },

    async completeOAuthLogin(accessToken) {
      if (!accessToken) {
        throw fromConnectError(new Error('missing token'));
      }

      storage.write({
        accessToken,
        account: { id: '', email: '' },
      });

      const session = await this.getSession();
      if (!session) {
        throw fromConnectError(new Error('session unavailable'));
      }
      return session;
    },

    async requestPasswordReset(email) {
      try {
        await client.requestPasswordReset({ email: email.trim() });
      } catch (err) {
        throw fromConnectError(err);
      }
    },

    async resetPassword(token, newPassword) {
      try {
        await client.resetPassword({ token, newPassword });
      } catch (err) {
        throw fromConnectError(err);
      }
    },

    async getVerificationStatus(email) {
      try {
        const response = await client.getVerificationStatus({ email: email.trim() });
        return response.verified;
      } catch (err) {
        throw fromConnectError(err);
      }
    },

    async resendVerificationEmail(email) {
      try {
        await client.resendVerificationEmail({ email: email.trim() });
      } catch (err) {
        throw fromConnectError(err);
      }
    },
  };
}
