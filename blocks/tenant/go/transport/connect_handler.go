package transport

import (
	"context"

	"connectrpc.com/connect"

	tenantv1 "github.com/emersonary/appkit/tenants/gen/tenant/v1"
)

type connectTenantService struct {
	inner *TenantServer
}

func newConnectTenantService(inner *TenantServer) *connectTenantService {
	return &connectTenantService{inner: inner}
}

func (h *connectTenantService) CreateTenant(
	ctx context.Context,
	req *connect.Request[tenantv1.CreateTenantRequest],
) (*connect.Response[tenantv1.TenantResponse], error) {
	resp, err := h.inner.CreateTenant(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectTenantService) ListMyTenants(
	ctx context.Context,
	req *connect.Request[tenantv1.ListMyTenantsRequest],
) (*connect.Response[tenantv1.ListMyTenantsResponse], error) {
	resp, err := h.inner.ListMyTenants(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectTenantService) GetTenant(
	ctx context.Context,
	req *connect.Request[tenantv1.GetTenantRequest],
) (*connect.Response[tenantv1.TenantResponse], error) {
	resp, err := h.inner.GetTenant(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectTenantService) InviteMember(
	ctx context.Context,
	req *connect.Request[tenantv1.InviteMemberRequest],
) (*connect.Response[tenantv1.InviteMemberResponse], error) {
	resp, err := h.inner.InviteMember(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectTenantService) AcceptInvite(
	ctx context.Context,
	req *connect.Request[tenantv1.AcceptInviteRequest],
) (*connect.Response[tenantv1.TenantMembershipResponse], error) {
	resp, err := h.inner.AcceptInvite(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
