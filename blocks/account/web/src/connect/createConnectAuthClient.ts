import { GetSessionRequest, LogoutRequest } from '../gen/v1/account_pb';
import { AuthService } from '../gen/v1/account_connect';
import type { AccountSession, AccountsAuthClient } from '../types';
import { fromConnectError } from './connectErrors';
import { createAccountsAuthClient } from './transport';
import type { AccountsSessionStorage } from './storage';

export type CreateConnectAuthClientOptions = {
  baseUrl?: string;
  storage: AccountsSessionStorage;
};

function toSession(response: {
  accessToken: string;
  user?: {
    id: string;
    email: string;
    firstName?: string;
    lastName?: string;
    avatarUrl?: string;
    isAdmin?: boolean;
  };
}): AccountSession {
  return {
    accessToken: response.accessToken,
    user: {
      id: response.user?.id ?? '',
      email: response.user?.email ?? '',
      firstName: response.user?.firstName || undefined,
      lastName: response.user?.lastName || undefined,
      avatarUrl: response.user?.avatarUrl || undefined,
      isAdmin: response.user?.isAdmin ?? false,
    },
  };
}

function isLegacyMockToken(token: string): boolean {
  return token.startsWith('mock-');
}

export function createConnectAuthClient({
  baseUrl = '',
  storage,
}: CreateConnectAuthClientOptions): AccountsAuthClient {
  const client = createAccountsAuthClient(AuthService, { baseUrl, storage });

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

    async register(email, password, firstName, lastName) {
      try {
        const response = await client.register({
          email: email.trim(),
          password,
          firstName: firstName.trim(),
          lastName: lastName?.trim() ?? '',
        });
        if (response.verificationRequired) {
          storage.write(null);
          return {
            verificationRequired: true,
            email: response.session?.user?.email ?? email.trim(),
          };
        }
        const session = toSession(response.session!);
        storage.write(session);
        return {
          verificationRequired: false,
          email: session.user.email,
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
        user: { id: '', email: '' },
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
