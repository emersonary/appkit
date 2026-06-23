package transport

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/emersonary/appkit/apperror"
	"github.com/emersonary/appkit/menu"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	var appErr apperror.Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case menu.ErrInvalidSetup.Code, menu.ErrPermissionsRequired.Code:
			return status.Error(codes.InvalidArgument, appErr.Error())
		case menu.ErrUnauthenticated.Code:
			return status.Error(codes.Unauthenticated, appErr.Error())
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
		case menu.ErrInvalidSetup.Code, menu.ErrPermissionsRequired.Code:
			return connect.NewError(connect.CodeInvalidArgument, appErr)
		case menu.ErrUnauthenticated.Code:
			return connect.NewError(connect.CodeUnauthenticated, appErr)
		}
	}

	return connect.NewError(connect.CodeInternal, errors.New("internal error"))
}
