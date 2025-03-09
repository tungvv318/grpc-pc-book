gen:
	# Generate Go code from .proto files
	protoc --proto_path=proto proto/*.proto --go_out=pb --go-grpc_out=pb

clean:
	# Remove generated protobuf files
	rm -f pb/*.go 

server:
	# Run the gRPC server
	go run cmd/server/main.go -port 8080

client:
	# Run the gRPC client
	go run cmd/client/main.go -address localhost:8080

test:
	# Run all tests with coverage and race detection
	go test -cover -race ./...
