package transport

import (
	"context"

	accountv1 "github.com/emersonary/appkit/accounts/gen/account/v1"
	"github.com/emersonary/appkit/accounts"
)

type AccountServer struct {
	accountv1.UnimplementedAccountServiceServer
	svc *accounts.Service
}

func NewAccountServer(svc *accounts.Service) *AccountServer {
	return &AccountServer{svc: svc}
}

func (s *AccountServer) Service() *accounts.Service {
	return s.svc
}

func (s *AccountServer) Login(ctx context.Context, req *accountv1.LoginRequest) (*accountv1.SessionResponse, error) {
	session, err := s.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return ToProtoSession(session), nil
}

func (s *AccountServer) AdminLogin(ctx context.Context, req *accountv1.LoginRequest) (*accountv1.SessionResponse, error) {
	session, err := s.svc.AdminLogin(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return ToProtoSession(session), nil
}

func (s *AccountServer) Register(ctx context.Context, req *accountv1.RegisterRequest) (*accountv1.RegisterResponse, error) {
	var firstName *string
	if name := req.GetFirstName(); name != "" {
		firstName = &name
	}
	var lastName *string
	if name := req.GetLastName(); name != "" {
		lastName = &name
	}
	var language *string
	if code := req.GetLanguage(); code != "" {
		language = &code
	}
	result, err := s.svc.Register(ctx, req.GetEmail(), req.GetPassword(), firstName, lastName, language)
	if err != nil {
		return nil, MapGRPCError(err)
	}

	resp := &accountv1.RegisterResponse{
		VerificationRequired: result.VerificationRequired,
	}
	if result.Session != nil {
		resp.Session = ToProtoSession(*result.Session)
	} else {
		resp.Session = &accountv1.SessionResponse{
			Account: ToProtoAccount(result.Account),
		}
	}
	return resp, nil
}

func (s *AccountServer) GetSession(ctx context.Context, _ *accountv1.GetSessionRequest) (*accountv1.SessionResponse, error) {
	token, err := BearerToken(ctx)
	if err != nil {
		return nil, err
	}
	session, err := s.svc.SessionFromToken(ctx, token)
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return ToProtoSession(session), nil
}

func (s *AccountServer) Logout(_ context.Context, _ *accountv1.LogoutRequest) (*accountv1.LogoutResponse, error) {
	return &accountv1.LogoutResponse{}, nil
}

func (s *AccountServer) GetVerificationStatus(ctx context.Context, req *accountv1.GetVerificationStatusRequest) (*accountv1.GetVerificationStatusResponse, error) {
	verified, err := s.svc.IsEmailVerified(ctx, req.GetEmail())
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &accountv1.GetVerificationStatusResponse{Verified: verified}, nil
}

func (s *AccountServer) ResendVerificationEmail(ctx context.Context, req *accountv1.ResendVerificationEmailRequest) (*accountv1.ResendVerificationEmailResponse, error) {
	if err := s.svc.ResendVerificationEmail(ctx, req.GetEmail()); err != nil {
		return nil, MapGRPCError(err)
	}
	return &accountv1.ResendVerificationEmailResponse{}, nil
}

func (s *AccountServer) RequestPasswordReset(ctx context.Context, req *accountv1.RequestPasswordResetRequest) (*accountv1.RequestPasswordResetResponse, error) {
	if err := s.svc.RequestPasswordReset(ctx, req.GetEmail()); err != nil {
		return nil, MapGRPCError(err)
	}
	return &accountv1.RequestPasswordResetResponse{}, nil
}

func (s *AccountServer) ResetPassword(ctx context.Context, req *accountv1.ResetPasswordRequest) (*accountv1.ResetPasswordResponse, error) {
	if err := s.svc.ResetPassword(ctx, req.GetToken(), req.GetNewPassword()); err != nil {
		return nil, MapGRPCError(err)
	}
	return &accountv1.ResetPasswordResponse{}, nil
}
