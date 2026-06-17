import { createContext, useContext } from 'react';
import type { TenantsClassNames, TenantsUILabels } from './types';

export type TenantsConfigContextValue = {
  labels: TenantsUILabels;
  classNames: TenantsClassNames;
};

export const defaultLabels: TenantsUILabels = {
  createTenantTitle: 'Create organization',
  tenantName: 'Organization name',
  tenantSlug: 'URL slug',
  tenantTimezone: 'Timezone',
  createTenant: 'Create organization',
  switchTenant: 'Switch organization',
  inviteEmail: 'Email',
  inviteRole: 'Role',
  sendInvite: 'Send invite',
  acceptInvite: 'Accept invite',
  inviteToken: 'Invite token',
};

export const TenantsConfigContext = createContext<TenantsConfigContextValue>({
  labels: defaultLabels,
  classNames: {},
});

export function useTenantsConfig() {
  return useContext(TenantsConfigContext);
}
