package transport

import (
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/emersonary/appkit/apperror"
	"github.com/emersonary/appkit/docupload"
)

func MapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	var appErr apperror.Error
	if errors.As(err, &appErr) {
		return status.Error(grpcCode(appErr.Code), appErr.Error())
	}
	if _, ok := status.FromError(err); ok {
		return err
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
		return connect.NewError(connectCode(appErr.Code), appErr)
	}
	return connect.NewError(connect.CodeInternal, errors.New("internal error"))
}

func grpcCode(code string) codes.Code {
	switch code {
	case docupload.ErrInvalidArgument.Code, docupload.ErrEmptyFile.Code, docupload.ErrTooLarge.Code, docupload.ErrChecksum.Code:
		return codes.InvalidArgument
	case docupload.ErrUnauthenticated.Code:
		return codes.Unauthenticated
	case docupload.ErrNotFound.Code:
		return codes.NotFound
	case docupload.ErrExpired.Code, docupload.ErrIncomplete.Code:
		return codes.FailedPrecondition
	case docupload.ErrConflict.Code:
		return codes.AlreadyExists
	default:
		return codes.Internal
	}
}

func connectCode(code string) connect.Code {
	switch code {
	case docupload.ErrInvalidArgument.Code, docupload.ErrEmptyFile.Code, docupload.ErrTooLarge.Code, docupload.ErrChecksum.Code:
		return connect.CodeInvalidArgument
	case docupload.ErrUnauthenticated.Code:
		return connect.CodeUnauthenticated
	case docupload.ErrNotFound.Code:
		return connect.CodeNotFound
	case docupload.ErrExpired.Code, docupload.ErrIncomplete.Code:
		return connect.CodeFailedPrecondition
	case docupload.ErrConflict.Code:
		return connect.CodeAlreadyExists
	default:
		return connect.CodeInternal
	}
}
