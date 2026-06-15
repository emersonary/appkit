PROTO_DIR := proto
AUTH_PROTO := $(PROTO_DIR)/auth/v1/auth.proto
GO_MODULE := github.com/emersonary/appkit
GO_DIR := go

.PHONY: proto proto-go proto-ts proto-install test test-go test-ts

proto: proto-go proto-ts

proto-go:
	protoc -I $(PROTO_DIR) \
		--go_out=$(GO_DIR) --go_opt=module=$(GO_MODULE) \
		--go-grpc_out=$(GO_DIR) --go-grpc_opt=module=$(GO_MODULE) \
		--connect-go_out=$(GO_DIR) --connect-go_opt=module=$(GO_MODULE) \
		$(AUTH_PROTO)

proto-ts:
	cd web/accounts && npm run generate:auth-proto

proto-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	cd web/accounts && npm install --save-dev @bufbuild/protoc-gen-es @connectrpc/protoc-gen-connect-es

test: test-go test-ts

test-go:
	cd $(GO_DIR) && go test ./...

test-ts:
	cd web/accounts && npm run test
