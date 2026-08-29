package transport

import (
	"context"

	"connectrpc.com/connect"

	accountv1 "github.com/emersonary/appkit/accounts/gen/account/v1"
)

type connectAccountService struct {
	inner *AccountServer
}

func newConnectAccountService(inner *AccountServer) *connectAccountService {
	return &connectAccountService{inner: inner}
}

func (h *connectAccountService) Login(
	ctx context.Context,
	req *connect.Request[accountv1.LoginRequest],
) (*connect.Response[accountv1.SessionResponse], error) {
	resp, err := h.inner.Login(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectAccountService) AdminLogin(
	ctx context.Context,
	req *connect.Request[accountv1.LoginRequest],
) (*connect.Response[accountv1.SessionResponse], error) {
	resp, err := h.inner.AdminLogin(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectAccountService) Register(
	ctx context.Context,
	req *connect.Request[accountv1.RegisterRequest],
) (*connect.Response[accountv1.RegisterResponse], error) {
	resp, err := h.inner.Register(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectAccountService) GetSession(
	ctx context.Context,
	req *connect.Request[accountv1.GetSessionRequest],
) (*connect.Response[accountv1.SessionResponse], error) {
	ctx = ContextWithAuthorizationHeader(ctx, req.Header().Get("Authorization"))
	resp, err := h.inner.GetSession(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectAccountService) Logout(
	ctx context.Context,
	req *connect.Request[accountv1.LogoutRequest],
) (*connect.Response[accountv1.LogoutResponse], error) {
	resp, err := h.inner.Logout(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectAccountService) GetVerificationStatus(
	ctx context.Context,
	req *connect.Request[accountv1.GetVerificationStatusRequest],
) (*connect.Response[accountv1.GetVerificationStatusResponse], error) {
	resp, err := h.inner.GetVerificationStatus(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectAccountService) ResendVerificationEmail(
	ctx context.Context,
	req *connect.Request[accountv1.ResendVerificationEmailRequest],
) (*connect.Response[accountv1.ResendVerificationEmailResponse], error) {
	resp, err := h.inner.ResendVerificationEmail(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectAccountService) RequestPasswordReset(
	ctx context.Context,
	req *connect.Request[accountv1.RequestPasswordResetRequest],
) (*connect.Response[accountv1.RequestPasswordResetResponse], error) {
	resp, err := h.inner.RequestPasswordReset(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectAccountService) ResetPassword(
	ctx context.Context,
	req *connect.Request[accountv1.ResetPasswordRequest],
) (*connect.Response[accountv1.ResetPasswordResponse], error) {
	resp, err := h.inner.ResetPassword(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
