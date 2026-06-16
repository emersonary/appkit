package transport

import (
	accountv1 "github.com/emersonary/appkit/accounts/gen/account/v1"
	"github.com/emersonary/appkit/accounts"
)

func ToProtoAccount(account accounts.Account) *accountv1.Account {
	pb := &accountv1.Account{
		Id:            account.ID,
		Email:         account.Email,
		EmailVerified: account.EmailVerified(),
		IsAdmin:       account.IsAdmin,
	}
	if account.FirstName != nil {
		pb.FirstName = *account.FirstName
	}
	if account.LastName != nil {
		pb.LastName = *account.LastName
	}
	if account.AvatarURL != nil {
		pb.AvatarUrl = *account.AvatarURL
	}
	return pb
}

func ToProtoSession(session accounts.Session) *accountv1.SessionResponse {
	return &accountv1.SessionResponse{
		AccessToken: session.AccessToken,
		Account:     ToProtoAccount(session.Account),
	}
}
