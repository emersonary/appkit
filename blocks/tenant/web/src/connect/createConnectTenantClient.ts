import { TenantService } from '../gen/v1/tenant_connect';
import type { Tenant, TenantMembership, TenantsClient } from '../types';
import { createTenantsClient } from './transport';
import type { TenantsSessionStorage } from '../types';

export type CreateConnectTenantClientOptions = {
  baseUrl?: string;
  storage: TenantsSessionStorage;
};

function toTenant(t: {
  id: string;
  slug: string;
  name: string;
  timezone: string;
  status: string;
}): Tenant {
  return {
    id: t.id,
    slug: t.slug,
    name: t.name,
    timezone: t.timezone,
    status: t.status,
  };
}

function toMembership(m: {
  tenant?: { id: string; slug: string; name: string; timezone: string; status: string };
  role: string;
}): TenantMembership {
  return {
    tenant: toTenant(m.tenant ?? { id: '', slug: '', name: '', timezone: 'UTC', status: 'trial' }),
    role: m.role,
  };
}

export function createConnectTenantClient({
  baseUrl = '',
  storage,
}: CreateConnectTenantClientOptions): TenantsClient {
  const client = createTenantsClient(TenantService, { baseUrl, storage });

  return {
    async createTenant(name, slug, timezone = 'UTC') {
      const response = await client.createTenant({ name, slug, timezone });
      return toTenant(response.tenant!);
    },
    async listMyTenants() {
      const response = await client.listMyTenants({});
      return (response.memberships ?? []).map(toMembership);
    },
    async getTenant(tenantId) {
      const response = await client.getTenant({ tenantId });
      return toTenant(response.tenant!);
    },
    async inviteMember(tenantId, email, role = 'staff') {
      const response = await client.inviteMember({ tenantId, email, role });
      return response.inviteToken;
    },
    async acceptInvite(inviteToken) {
      const response = await client.acceptInvite({ inviteToken });
      return toMembership(response.membership!);
    },
  };
}
