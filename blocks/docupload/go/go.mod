module github.com/emersonary/appkit/docupload

go 1.26.3

require (
	connectrpc.com/connect v1.20.0
	github.com/alicebob/miniredis/v2 v2.35.0
	github.com/emersonary/appkit v0.0.0
	github.com/redis/go-redis/v9 v9.21.0
	go.uber.org/zap v1.28.0
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260420184626-e10c466a9529 // indirect
)

replace github.com/emersonary/appkit => ../../../go
