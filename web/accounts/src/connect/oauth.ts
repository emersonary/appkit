import { accountsHttpUrl, isAccountsApiConfigured } from './transport';

export function googleAccountsLoginUrl(baseUrl?: string): string {
  if (!isAccountsApiConfigured(baseUrl)) {
    throw new Error('Accounts API is not configured');
  }
  return accountsHttpUrl('/auth/google', baseUrl);
}
