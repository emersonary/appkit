import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  createSocialConnectionClient,
  type SocialConnectionLabels,
  type SocialConnectionStatus,
} from './api';

export type UseSocialConnectionOptions = {
  platform: string;
  language: string;
  apiBaseUrl: string;
  getAccessToken: () => string | undefined;
};

export function useSocialConnection(options: UseSocialConnectionOptions) {
  const client = useMemo(
    () =>
      createSocialConnectionClient({
        platform: options.platform,
        language: options.language,
        apiBaseUrl: options.apiBaseUrl,
        getAccessToken: options.getAccessToken,
      }),
    [options.apiBaseUrl, options.getAccessToken, options.language, options.platform],
  );

  const [status, setStatus] = useState<SocialConnectionStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const next = await client.fetchStatus();
      setStatus(next);
      setError(null);
    } catch {
      setStatus(null);
      setError('load');
    } finally {
      setLoading(false);
    }
  }, [client]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function connect() {
    setBusy(true);
    setError(null);
    try {
      client.startConnect();
    } catch {
      setError('connect');
      setBusy(false);
    }
  }

  async function disconnect() {
    setBusy(true);
    setError(null);
    try {
      await client.disconnect();
      await refresh();
    } catch {
      setError('disconnect');
    } finally {
      setBusy(false);
    }
  }

  return { status, loading, busy, error, refresh, connect, disconnect };
}

export type SocialConnectionCardProps = {
  platform: string;
  language: string;
  apiBaseUrl: string;
  getAccessToken: () => string | undefined;
  labels?: SocialConnectionLabels;
  className?: string;
  compact?: boolean;
  /** When true, shows the disconnect control (hidden by default). */
  allowDisconnect?: boolean;
};

const defaultLabels = (platformLabel: string): Required<SocialConnectionLabels> => ({
  title: platformLabel,
  loading: `Checking ${platformLabel} connection…`,
  connectHint: `Connect ${platformLabel} to publish automatically.`,
  notConfigured: `Set ${platformLabel} OAuth credentials (client ID and secret) on the server before connecting.`,
  connected: 'Connected',
  expiresIn: (days) => `expires in ${days} day(s)`,
  expired: `${platformLabel} token expired. Reconnect to publish again.`,
  expiringSoon: 'Token expires soon. Reconnect to avoid publish failures.',
  connectButton: `Connect ${platformLabel}`,
  connectingButton: 'Connecting…',
  reconnectButton: `Reconnect ${platformLabel}`,
  disconnectButton: 'Disconnect',
  loadError: `Could not load ${platformLabel} status.`,
  connectError: `Could not start ${platformLabel} connection.`,
  disconnectError: `Could not disconnect ${platformLabel}.`,
});

export function SocialConnectionCard({
  platform,
  language,
  apiBaseUrl,
  getAccessToken,
  labels,
  className,
  compact = false,
  allowDisconnect = false,
}: SocialConnectionCardProps) {
  const platformLabel = labels?.title ?? platform.toUpperCase();
  const text = { ...defaultLabels(platformLabel), ...labels };
  const { status, loading, busy, error, connect, disconnect } = useSocialConnection({
    platform,
    language,
    apiBaseUrl,
    getAccessToken,
  });

  if (loading) {
    return <p className={className}>{text.loading}</p>;
  }

  const connected = status?.connected === true;
  const expired = status?.expired === true;
  const oauthConfigured = status?.oauth_configured !== false;
  const expiringSoon = connected && (status?.days_left ?? 0) <= 7;

  let errorMessage: string | null = null;
  if (error === 'load') errorMessage = text.loadError;
  if (error === 'connect') errorMessage = text.connectError;
  if (error === 'disconnect') errorMessage = text.disconnectError;

  if (compact) {
    return (
      <div className={`social-connection-inline ${className ?? ''}`.trim()}>
        {errorMessage && <span className="text-danger social-connection-inline__text">{errorMessage}</span>}
        {!oauthConfigured && <span className="text-muted social-connection-inline__text">{text.notConfigured}</span>}
        {connected && !expired && (
          <span className="social-connection-card__status social-connection-card__status--ok social-connection-inline__text">
            {text.connected}
            {status?.expires_at ? ` · ${text.expiresIn(status.days_left)}` : ''}
          </span>
        )}
        {expired && <span className="text-danger social-connection-inline__text">{text.expired}</span>}
        {expiringSoon && <span className="text-warning social-connection-inline__text">{text.expiringSoon}</span>}
        <div className="social-connection-inline__actions">
          {(!connected || expired || expiringSoon) && oauthConfigured && (
            <button type="button" className="btn btn-outline btn-sm" disabled={busy} onClick={() => void connect()}>
              {busy ? text.connectingButton : connected ? text.reconnectButton : text.connectButton}
            </button>
          )}
          {allowDisconnect && connected && (
            <button type="button" className="btn btn-outline btn-sm" disabled={busy} onClick={() => void disconnect()}>
              {text.disconnectButton}
            </button>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className={`card ${className ?? ''}`.trim()}>
      <h2 style={{ marginTop: 0, fontSize: '1rem' }}>{text.title}</h2>
      {errorMessage && <div className="alert alert-warning">{errorMessage}</div>}
      {!oauthConfigured && (
        <p className="alert alert-warning">{text.notConfigured}</p>
      )}
      {connected && !expired && (
        <p className="social-connection-card__status social-connection-card__status--ok">
          {text.connected}
          {status?.expires_at && (
            <>
              {' '}
              · <strong>{text.expiresIn(status.days_left)}</strong>
            </>
          )}
        </p>
      )}
      {expired && <p className="text-danger">{text.expired}</p>}
      {!connected && !expired && <p className="text-muted">{text.connectHint}</p>}
      {oauthConfigured && status?.redirect_uri && (
        <p className="text-muted" style={{ fontSize: '0.85rem', marginBottom: '0.5rem' }}>
          LinkedIn redirect URI: <code>{status.redirect_uri}</code>
        </p>
      )}
      {expiringSoon && (
        <p className="alert alert-warning" style={{ marginBottom: '0.75rem' }}>
          {text.expiringSoon}
        </p>
      )}
      <div className="form-actions" style={{ marginTop: '0.75rem' }}>
        {(!connected || expired || expiringSoon) && oauthConfigured && (
          <button type="button" className="btn btn-primary" disabled={busy} onClick={() => void connect()}>
            {busy ? text.connectingButton : connected ? text.reconnectButton : text.connectButton}
          </button>
        )}
        {allowDisconnect && connected && (
          <button type="button" className="btn btn-outline" disabled={busy} onClick={() => void disconnect()}>
            {text.disconnectButton}
          </button>
        )}
      </div>
    </div>
  );
}
