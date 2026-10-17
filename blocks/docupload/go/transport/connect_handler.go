package transport

import (
	"context"

	"connectrpc.com/connect"

	docuploadv1 "github.com/emersonary/appkit/docupload/gen/docupload/v1"
)

type connectService struct {
	inner *Server
}

func newConnectService(inner *Server) *connectService {
	return &connectService{inner: inner}
}

func (h *connectService) StartUpload(ctx context.Context, req *connect.Request[docuploadv1.StartUploadRequest]) (*connect.Response[docuploadv1.StartUploadResponse], error) {
	resp, err := h.inner.StartUpload(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectService) UploadBlock(ctx context.Context, req *connect.Request[docuploadv1.UploadBlockRequest]) (*connect.Response[docuploadv1.UploadBlockResponse], error) {
	resp, err := h.inner.UploadBlock(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *connectService) CancelUpload(ctx context.Context, req *connect.Request[docuploadv1.CancelUploadRequest]) (*connect.Response[docuploadv1.CancelUploadResponse], error) {
	resp, err := h.inner.CancelUpload(ctx, req.Msg)
	if err != nil {
		return nil, ToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
