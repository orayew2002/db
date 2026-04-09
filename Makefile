run:
	@go run main.go

test:
	@go test ./... -v

cli:
	@go run cmd/cli/main.go
