import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { TenantsConfigContext, defaultLabels } from './context';
import type { Tenant, TenantMembership, TenantsProviderProps } from './types';

export type TenantsContextValue = {
  memberships: TenantMembership[];
  activeTenant: Tenant | null;
  activeTenantId: string | null;
  isLoading: boolean;
  isInitialized: boolean;
  refreshMemberships: () => Promise<void>;
  setActiveTenantId: (tenantId: string | null) => void;
  createTenant: (name: string, slug: string, timezone?: string) => Promise<Tenant>;
  inviteMember: (tenantId: string, email: string, role?: string) => Promise<string>;
  acceptInvite: (inviteToken: string) => Promise<TenantMembership>;
};

const TenantsCtx = createContext<TenantsContextValue | null>(null);

export function TenantsProvider({
  children,
  tenantClient,
  activeTenantId: controlledTenantId,
  onActiveTenantChange,
  labels,
  classNames,
}: TenantsProviderProps) {
  const [memberships, setMemberships] = useState<TenantMembership[]>([]);
  const [internalTenantId, setInternalTenantId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);

  const activeTenantId = controlledTenantId ?? internalTenantId;

  const setActiveTenantId = useCallback(
    (tenantId: string | null) => {
      if (onActiveTenantChange) {
        onActiveTenantChange(tenantId);
      } else {
        setInternalTenantId(tenantId);
      }
    },
    [onActiveTenantChange],
  );

  const configValue = useMemo(
    () => ({
      labels: { ...defaultLabels, ...labels },
      classNames: classNames ?? {},
    }),
    [labels, classNames],
  );

  const refreshMemberships = useCallback(async () => {
    setIsLoading(true);
    try {
      const next = await tenantClient.listMyTenants();
      setMemberships(next);
      if (next.length === 1 && !activeTenantId) {
        setActiveTenantId(next[0].tenant.id);
      }
    } finally {
      setIsLoading(false);
    }
  }, [tenantClient, activeTenantId, setActiveTenantId]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const next = await tenantClient.listMyTenants();
        if (!cancelled) {
          setMemberships(next);
          if (next.length === 1 && !activeTenantId) {
            setActiveTenantId(next[0].tenant.id);
          }
        }
      } finally {
        if (!cancelled) setIsInitialized(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [tenantClient, activeTenantId, setActiveTenantId]);

  const createTenant = useCallback(
    async (name: string, slug: string, timezone?: string) => {
      setIsLoading(true);
      try {
        const tenant = await tenantClient.createTenant(name, slug, timezone);
        await refreshMemberships();
        setActiveTenantId(tenant.id);
        return tenant;
      } finally {
        setIsLoading(false);
      }
    },
    [tenantClient, refreshMemberships, setActiveTenantId],
  );

  const inviteMember = useCallback(
    async (tenantId: string, email: string, role?: string) => {
      setIsLoading(true);
      try {
        return await tenantClient.inviteMember(tenantId, email, role);
      } finally {
        setIsLoading(false);
      }
    },
    [tenantClient],
  );

  const acceptInvite = useCallback(
    async (inviteToken: string) => {
      setIsLoading(true);
      try {
        const membership = await tenantClient.acceptInvite(inviteToken);
        await refreshMemberships();
        setActiveTenantId(membership.tenant.id);
        return membership;
      } finally {
        setIsLoading(false);
      }
    },
    [tenantClient, refreshMemberships, setActiveTenantId],
  );

  const activeTenant = useMemo(
    () => memberships.find((m) => m.tenant.id === activeTenantId)?.tenant ?? null,
    [memberships, activeTenantId],
  );

  const value = useMemo(
    () => ({
      memberships,
      activeTenant,
      activeTenantId,
      isLoading,
      isInitialized,
      refreshMemberships,
      setActiveTenantId,
      createTenant,
      inviteMember,
      acceptInvite,
    }),
    [
      memberships,
      activeTenant,
      activeTenantId,
      isLoading,
      isInitialized,
      refreshMemberships,
      setActiveTenantId,
      createTenant,
      inviteMember,
      acceptInvite,
    ],
  );

  return (
    <TenantsConfigContext.Provider value={configValue}>
      <TenantsCtx.Provider value={value}>{children}</TenantsCtx.Provider>
    </TenantsConfigContext.Provider>
  );
}

export function useTenants() {
  const ctx = useContext(TenantsCtx);
  if (!ctx) {
    throw new Error('useTenants must be used within TenantsProvider');
  }
  return ctx;
}
