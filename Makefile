PROTO_DIR := proto
AUTH_PROTO := $(PROTO_DIR)/auth/v1/auth.proto
GO_MODULE := github.com/emersonary/appkit
ACCOUNTS_GO_MODULE := github.com/emersonary/appkit/accounts
GO_DIR := go
ACCOUNTS_GO_DIR := blocks/account/go
ACCOUNTS_WEB_DIR := blocks/account/web

.PHONY: proto proto-go proto-ts proto-install test test-go test-ts test-accounts

proto: proto-go proto-ts

proto-go:
	protoc -I $(PROTO_DIR) \
		--go_out=$(ACCOUNTS_GO_DIR) --go_opt=module=$(ACCOUNTS_GO_MODULE) \
		--go-grpc_out=$(ACCOUNTS_GO_DIR) --go-grpc_opt=module=$(ACCOUNTS_GO_MODULE) \
		--connect-go_out=$(ACCOUNTS_GO_DIR) --connect-go_opt=module=$(ACCOUNTS_GO_MODULE) \
		$(AUTH_PROTO)

proto-ts:
	cd $(ACCOUNTS_WEB_DIR) && npm run generate:auth-proto

proto-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	cd $(ACCOUNTS_WEB_DIR) && npm install --save-dev @bufbuild/protoc-gen-es @connectrpc/protoc-gen-connect-es

test: test-go test-ts

test-go:
	cd $(GO_DIR) && go test ./...
	cd $(ACCOUNTS_GO_DIR) && go test ./...

test-ts:
	cd $(ACCOUNTS_WEB_DIR) && npm run test
