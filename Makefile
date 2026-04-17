test:
	@go test ./... -v

cli:
	@go run cmd/cli/main.go

dump:
	@go run cmd/seed/main.go

pprof:
	@echo "Running benchmark and saving CPU profile to cpu.prof..."
	@go test ./src/db -bench=Insert -benchmem -cpuprofile=cpu.prof
	@go tool pprof cpu.prof

proto-gen:
	protoc --go_out=. ./src/proto/*.proto
