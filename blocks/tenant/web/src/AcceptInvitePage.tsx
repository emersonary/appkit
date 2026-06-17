import { type FormEvent, useState } from 'react';
import { useTenantsConfig } from './context';
import { useTenants } from './TenantsProvider';

export type AcceptInvitePageProps = {
  initialToken?: string;
  onAccepted?: () => void;
};

export function AcceptInvitePage({ initialToken = '', onAccepted }: AcceptInvitePageProps) {
  const { labels, classNames } = useTenantsConfig();
  const { acceptInvite, isLoading } = useTenants();
  const [token, setToken] = useState(initialToken);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      await acceptInvite(token);
      onAccepted?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to accept invite');
    }
  }

  return (
    <form className={classNames.form} onSubmit={onSubmit}>
      <label className={classNames.label}>
        {labels.inviteToken}
        <input
          className={classNames.input}
          value={token}
          onChange={(e) => setToken(e.target.value)}
          required
        />
      </label>
      {error ? <p className={classNames.error}>{error}</p> : null}
      <button className={classNames.button} type="submit" disabled={isLoading}>
        {labels.acceptInvite}
      </button>
    </form>
  );
}
