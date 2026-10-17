package transport

import (
	"context"

	"github.com/emersonary/appkit/docupload"
	docuploadv1 "github.com/emersonary/appkit/docupload/gen/docupload/v1"
)

// Server adapts the domain service to gRPC.
type Server struct {
	docuploadv1.UnimplementedDocUploadServiceServer
	svc *docupload.Service
}

func NewServer(svc *docupload.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) StartUpload(ctx context.Context, req *docuploadv1.StartUploadRequest) (*docuploadv1.StartUploadResponse, error) {
	accountID, err := accountID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := s.svc.StartUpload(ctx, accountID, docupload.StartRequest{
		FileID:         req.GetFileId(),
		FileName:       req.GetFileName(),
		FileUnixMs:     req.GetFileUnixMs(),
		OriginalSize:   req.GetOriginalSize(),
		CompressedMD5:  req.GetCompressedMd5(),
		CompressedSize: req.GetCompressedSize(),
		ExpiresUnixMs:  req.GetExpiresUnixMs(),
		ContentType:    req.GetContentType(),
		Source:         req.GetSource(),
	})
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &docuploadv1.StartUploadResponse{
		BlockSize:     result.BlockSize,
		PendingBlocks: result.Pending,
		MaxThreads:    result.MaxThreads,
		ExpiresUnixMs: result.ExpiresUnixMs,
	}, nil
}

func (s *Server) UploadBlock(ctx context.Context, req *docuploadv1.UploadBlockRequest) (*docuploadv1.UploadBlockResponse, error) {
	accountID, err := accountID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := s.svc.UploadBlock(ctx, accountID, docupload.BlockRequest{
		FileID:   req.GetFileId(),
		Index:    req.GetIndex(),
		Payload:  req.GetPayload(),
		BlockMD5: req.GetBlockMd5(),
	})
	if err != nil {
		return nil, MapGRPCError(err)
	}
	return &docuploadv1.UploadBlockResponse{
		Success:       result.Success,
		PendingBlocks: result.Pending,
	}, nil
}

func (s *Server) CancelUpload(ctx context.Context, req *docuploadv1.CancelUploadRequest) (*docuploadv1.CancelUploadResponse, error) {
	accountID, err := accountID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.svc.CancelUpload(ctx, accountID, req.GetFileId()); err != nil {
		return nil, MapGRPCError(err)
	}
	return &docuploadv1.CancelUploadResponse{}, nil
}

func accountID(ctx context.Context) (string, error) {
	id, ok := AccountIDFromContext(ctx)
	if !ok {
		return "", MapGRPCError(docupload.ErrUnauthenticated)
	}
	return id, nil
}
