package transport

import (
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/emersonary/appkit/apperror"
	"github.com/emersonary/appkit/tenants"
)

func MapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	var appErr apperror.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case tenants.ErrInvalidArgument.Code:
			return status.Error(codes.InvalidArgument, appErr.Error())
		case tenants.ErrUnauthenticated.Code, tenants.ErrInvalidToken.Code:
			return status.Error(codes.Unauthenticated, appErr.Error())
		case tenants.ErrForbidden.Code, tenants.ErrNotFound.Code:
			return status.Error(codes.PermissionDenied, appErr.Error())
		case tenants.ErrAlreadyExists.Code:
			return status.Error(codes.AlreadyExists, appErr.Error())
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
		case tenants.ErrInvalidArgument.Code:
			return connect.NewError(connect.CodeInvalidArgument, appErr)
		case tenants.ErrUnauthenticated.Code, tenants.ErrInvalidToken.Code:
			return connect.NewError(connect.CodeUnauthenticated, appErr)
		case tenants.ErrForbidden.Code, tenants.ErrNotFound.Code:
			return connect.NewError(connect.CodePermissionDenied, appErr)
		case tenants.ErrAlreadyExists.Code:
			return connect.NewError(connect.CodeAlreadyExists, appErr)
		}
	}

	return connect.NewError(connect.CodeInternal, errors.New("internal error"))
}
