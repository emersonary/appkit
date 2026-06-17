import { useTenantsConfig } from './context';
import { useTenants } from './TenantsProvider';

export function TenantSwitcher() {
  const { labels, classNames } = useTenantsConfig();
  const { memberships, activeTenantId, setActiveTenantId } = useTenants();

  if (memberships.length <= 1) {
    return null;
  }

  return (
    <label className={classNames.label}>
      {labels.switchTenant}
      <select
        className={classNames.select}
        value={activeTenantId ?? ''}
        onChange={(e) => setActiveTenantId(e.target.value || null)}
      >
        <option value="">Select organization</option>
        {memberships.map((m) => (
          <option key={m.tenant.id} value={m.tenant.id}>
            {m.tenant.name}
          </option>
        ))}
      </select>
    </label>
  );
}
