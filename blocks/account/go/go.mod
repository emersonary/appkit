module github.com/emersonary/appkit/accounts

go 1.26.3

require (
	connectrpc.com/connect v1.20.0
	github.com/emersonary/appkit v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	golang.org/x/crypto v0.50.0
	golang.org/x/oauth2 v0.36.0
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260420184626-e10c466a9529 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/emersonary/appkit => ../../../go
