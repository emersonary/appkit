package transport

import (
	"context"

	tenantv1 "github.com/emersonary/appkit/tenants/gen/tenant/v1"
	"github.com/emersonary/appkit/tenants"
)

type TenantServer struct {
	tenantv1.UnimplementedTenantServiceServer
	svc *tenants.Service
}

func NewTenantServer(svc *tenants.Service) *TenantServer {
	return &TenantServer{svc: svc}
}

func (s *TenantServer) CreateTenant(ctx context.Context, req *tenantv1.CreateTenantRequest) (*tenantv1.TenantResponse, error) {
	accountID, ok := AccountIDFromContext(ctx)
	if !ok {
		return nil, MapGRPCError(tenants.ErrUnauthenticated)
	}

	membership, err := s.svc.CreateTenant(ctx, accountID, req.GetName(), req.GetSlug(), req.GetTimezone())
	if err != nil {
		return nil, MapGRPCError(err)
	}

	return &tenantv1.TenantResponse{Tenant: toProtoTenant(membership.Tenant)}, nil
}

func (s *TenantServer) ListMyTenants(ctx context.Context, _ *tenantv1.ListMyTenantsRequest) (*tenantv1.ListMyTenantsResponse, error) {
	accountID, ok := AccountIDFromContext(ctx)
	if !ok {
		return nil, MapGRPCError(tenants.ErrUnauthenticated)
	}

	memberships, err := s.svc.ListMyTenants(ctx, accountID)
	if err != nil {
		return nil, MapGRPCError(err)
	}

	out := make([]*tenantv1.TenantMembership, 0, len(memberships))
	for _, m := range memberships {
		out = append(out, toProtoMembership(m))
	}
	return &tenantv1.ListMyTenantsResponse{Memberships: out}, nil
}

func (s *TenantServer) GetTenant(ctx context.Context, req *tenantv1.GetTenantRequest) (*tenantv1.TenantResponse, error) {
	accountID, ok := AccountIDFromContext(ctx)
	if !ok {
		return nil, MapGRPCError(tenants.ErrUnauthenticated)
	}

	tenant, err := s.svc.GetTenant(ctx, accountID, req.GetTenantId())
	if err != nil {
		return nil, MapGRPCError(err)
	}

	return &tenantv1.TenantResponse{Tenant: toProtoTenant(tenant)}, nil
}

func (s *TenantServer) InviteMember(ctx context.Context, req *tenantv1.InviteMemberRequest) (*tenantv1.InviteMemberResponse, error) {
	accountID, ok := AccountIDFromContext(ctx)
	if !ok {
		return nil, MapGRPCError(tenants.ErrUnauthenticated)
	}

	token, err := s.svc.InviteMember(ctx, accountID, req.GetTenantId(), req.GetEmail(), req.GetRole())
	if err != nil {
		return nil, MapGRPCError(err)
	}

	return &tenantv1.InviteMemberResponse{InviteToken: token}, nil
}

func (s *TenantServer) AcceptInvite(ctx context.Context, req *tenantv1.AcceptInviteRequest) (*tenantv1.TenantMembershipResponse, error) {
	accountID, ok := AccountIDFromContext(ctx)
	if !ok {
		return nil, MapGRPCError(tenants.ErrUnauthenticated)
	}

	membership, err := s.svc.AcceptInvite(ctx, accountID, req.GetInviteToken())
	if err != nil {
		return nil, MapGRPCError(err)
	}

	return &tenantv1.TenantMembershipResponse{Membership: toProtoMembership(membership)}, nil
}

func toProtoTenant(t tenants.Tenant) *tenantv1.Tenant {
	return &tenantv1.Tenant{
		Id:       t.ID,
		Slug:     t.Slug,
		Name:     t.Name,
		Timezone: t.Timezone,
		Status:   t.Status,
	}
}

func toProtoMembership(m tenants.Membership) *tenantv1.TenantMembership {
	return &tenantv1.TenantMembership{
		Tenant: toProtoTenant(m.Tenant),
		Role:   m.Role,
	}
}
