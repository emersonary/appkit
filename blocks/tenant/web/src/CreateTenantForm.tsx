import { type FormEvent, useState } from 'react';
import { useTenantsConfig } from './context';
import { useTenants } from './TenantsProvider';

export type CreateTenantFormProps = {
  defaultTimezone?: string;
};

export function CreateTenantForm({ defaultTimezone = 'UTC' }: CreateTenantFormProps) {
  const { labels, classNames } = useTenantsConfig();
  const { createTenant, isLoading } = useTenants();
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [timezone, setTimezone] = useState(defaultTimezone);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      await createTenant(name, slug, timezone);
      setName('');
      setSlug('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create organization');
    }
  }

  return (
    <form className={classNames.form} onSubmit={onSubmit}>
      <h2>{labels.createTenantTitle}</h2>
      <label className={classNames.label}>
        {labels.tenantName}
        <input
          className={classNames.input}
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </label>
      <label className={classNames.label}>
        {labels.tenantSlug}
        <input
          className={classNames.input}
          value={slug}
          onChange={(e) => setSlug(e.target.value.toLowerCase())}
          required
        />
      </label>
      <label className={classNames.label}>
        {labels.tenantTimezone}
        <input
          className={classNames.input}
          value={timezone}
          onChange={(e) => setTimezone(e.target.value)}
        />
      </label>
      {error ? <p className={classNames.error}>{error}</p> : null}
      <button className={classNames.button} type="submit" disabled={isLoading}>
        {labels.createTenant}
      </button>
    </form>
  );
}
