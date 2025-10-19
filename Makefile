# Generate protobuf libraries from .proto files
.PHONY: generate-proto-libs
generate-proto-libs:
	buf generate

# Validate (and formats) all .txtpb files in the examples/ directory
.PHONY: validate-examples
validate-examples:
	find ./examples -type f -name "*.txtpb" | xargs go run github.com/protocolbuffers/txtpbfmt/cmd/txtpbfmt@latest
	go run tools/validatetxtpb/main.go --dir=./examples
