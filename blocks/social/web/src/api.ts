export type SocialConnectionStatus = {
  oauth_configured: boolean;
  connected: boolean;
  expired: boolean;
  account_id?: string;
  expires_at?: string;
  days_left: number;
};

export type SocialConnectionLabels = {
  title?: string;
  loading?: string;
  connectHint?: string;
  notConfigured?: string;
  connected?: string;
  expiresIn?: (days: number) => string;
  expired?: string;
  expiringSoon?: string;
  connectButton?: string;
  connectingButton?: string;
  reconnectButton?: string;
  disconnectButton?: string;
  loadError?: string;
  connectError?: string;
  disconnectError?: string;
};

export type SocialConnectionClientOptions = {
  apiBaseUrl: string;
  platform: string;
  language: string;
  getAccessToken: () => string | undefined;
};

function normalizeBaseUrl(url: string): string {
  return url.replace(/\/$/, '');
}

function authHeaders(getAccessToken: () => string | undefined): HeadersInit {
  const token = getAccessToken();
  if (!token) {
    throw new Error('missing admin session');
  }
  return { Authorization: `Bearer ${token}` };
}

export function createSocialConnectionClient(options: SocialConnectionClientOptions) {
  const base = normalizeBaseUrl(options.apiBaseUrl);
  const platform = options.platform.trim().toLowerCase();
  const language = options.language.trim().toLowerCase();

  return {
    async fetchStatus(): Promise<SocialConnectionStatus> {
      const res = await fetch(`${base}/auth/social/${platform}/${language}/status`, {
        headers: authHeaders(options.getAccessToken),
      });
      if (!res.ok) {
        throw new Error('failed to load social connection status');
      }
      return res.json() as Promise<SocialConnectionStatus>;
    },

    startConnect(): void {
      const token = options.getAccessToken();
      if (!token) {
        throw new Error('missing admin session');
      }
      const url = new URL(`${base}/auth/social/${platform}/${language}/start`);
      url.searchParams.set('token', token);
      window.location.href = url.toString();
    },

    async disconnect(): Promise<void> {
      const res = await fetch(`${base}/auth/social/${platform}/${language}`, {
        method: 'DELETE',
        headers: authHeaders(options.getAccessToken),
      });
      if (!res.ok && res.status !== 204) {
        throw new Error('failed to disconnect social platform');
      }
    },
  };
}
