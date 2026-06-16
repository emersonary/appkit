package transport

import (
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/apperror"
)

func MapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	var appErr apperror.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case accounts.ErrInvalidArgument.Code:
			return status.Error(codes.InvalidArgument, appErr.Error())
		case accounts.ErrUnauthenticated.Code, accounts.ErrInvalidToken.Code, accounts.ErrNotFound.Code:
			return status.Error(codes.Unauthenticated, appErr.Error())
		case accounts.ErrAlreadyExists.Code:
			return status.Error(codes.AlreadyExists, appErr.Error())
		case accounts.ErrEmailNotVerified.Code:
			return status.Error(codes.FailedPrecondition, appErr.Error())
		}
	}
	return status.Error(codes.Internal, "internal error")
}

func ToConnectError(err error) error {
	if err == nil {
		return nil
	}

	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return err
	}

	if st, ok := status.FromError(err); ok {
		return connect.NewError(connect.Code(st.Code()), errors.New(st.Message()))
	}

	var appErr apperror.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case accounts.ErrInvalidArgument.Code:
			return connect.NewError(connect.CodeInvalidArgument, appErr)
		case accounts.ErrUnauthenticated.Code, accounts.ErrInvalidToken.Code, accounts.ErrNotFound.Code:
			return connect.NewError(connect.CodeUnauthenticated, appErr)
		case accounts.ErrAlreadyExists.Code:
			return connect.NewError(connect.CodeAlreadyExists, appErr)
		case accounts.ErrEmailNotVerified.Code:
			return connect.NewError(connect.CodeFailedPrecondition, appErr)
		}
	}

	return connect.NewError(connect.CodeInternal, errors.New("internal error"))
}
