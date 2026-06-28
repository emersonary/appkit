package transport

import (
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/grpc"

	blogv1 "github.com/emersonary/appkit/blog/gen/blog/v1"
	"github.com/emersonary/appkit/blog/gen/blog/v1/blogv1connect"
)

// Mount configures where blog transport is registered.
type Mount struct {
	HTTPMux        *http.ServeMux
	GRPCServer     *grpc.Server
	ConnectOptions []connect.HandlerOption
}

// Register mounts BlogService Connect (and optional gRPC) handlers.
func Register(svc blogv1connect.BlogServiceHandler, mount *Mount) string {
	if mount == nil {
		return ""
	}

	path := ""
	if mount.HTTPMux != nil {
		var handler http.Handler
		path, handler = blogv1connect.NewBlogServiceHandler(svc, mount.ConnectOptions...)
		mount.HTTPMux.Handle(path, handler)
	}

	if mount.GRPCServer != nil {
		if grpcSvc, ok := svc.(blogv1.BlogServiceServer); ok {
			blogv1.RegisterBlogServiceServer(mount.GRPCServer, grpcSvc)
		}
	}

	return path
}
