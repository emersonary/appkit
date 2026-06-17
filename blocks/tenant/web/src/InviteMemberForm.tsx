import { type FormEvent, useState } from 'react';
import { useTenantsConfig } from './context';
import { useTenants } from './TenantsProvider';

export type InviteMemberFormProps = {
  tenantId: string;
  onInvited?: (inviteToken: string) => void;
};

export function InviteMemberForm({ tenantId, onInvited }: InviteMemberFormProps) {
  const { labels, classNames } = useTenantsConfig();
  const { inviteMember, isLoading } = useTenants();
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('staff');
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      const token = await inviteMember(tenantId, email, role);
      onInvited?.(token);
      setEmail('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send invite');
    }
  }

  return (
    <form className={classNames.form} onSubmit={onSubmit}>
      <label className={classNames.label}>
        {labels.inviteEmail}
        <input
          className={classNames.input}
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />
      </label>
      <label className={classNames.label}>
        {labels.inviteRole}
        <select className={classNames.select} value={role} onChange={(e) => setRole(e.target.value)}>
          <option value="admin">admin</option>
          <option value="staff">staff</option>
          <option value="viewer">viewer</option>
        </select>
      </label>
      {error ? <p className={classNames.error}>{error}</p> : null}
      <button className={classNames.button} type="submit" disabled={isLoading}>
        {labels.sendInvite}
      </button>
    </form>
  );
}
