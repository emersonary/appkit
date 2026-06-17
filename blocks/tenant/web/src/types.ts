export type Tenant = {
  id: string;
  slug: string;
  name: string;
  timezone: string;
  status: string;
};

export type TenantMembership = {
  tenant: Tenant;
  role: string;
};

export type TenantsSessionStorage = {
  readAccessToken: () => string | null;
};

export type TenantsClient = {
  createTenant(name: string, slug: string, timezone?: string): Promise<Tenant>;
  listMyTenants(): Promise<TenantMembership[]>;
  getTenant(tenantId: string): Promise<Tenant>;
  inviteMember(tenantId: string, email: string, role?: string): Promise<string>;
  acceptInvite(inviteToken: string): Promise<TenantMembership>;
};

export type TenantsClassNames = {
  form?: string;
  field?: string;
  label?: string;
  input?: string;
  button?: string;
  select?: string;
  error?: string;
};

export type TenantsUILabels = {
  createTenantTitle?: string;
  tenantName?: string;
  tenantSlug?: string;
  tenantTimezone?: string;
  createTenant?: string;
  switchTenant?: string;
  inviteEmail?: string;
  inviteRole?: string;
  sendInvite?: string;
  acceptInvite?: string;
  inviteToken?: string;
};

export type TenantsProviderProps = {
  children: React.ReactNode;
  tenantClient: TenantsClient;
  activeTenantId?: string | null;
  onActiveTenantChange?: (tenantId: string | null) => void;
  labels?: TenantsUILabels;
  classNames?: TenantsClassNames;
};
