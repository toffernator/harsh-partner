# Runs the protoc command to compile protofiles
compile-proto:
	@mkdir -p api/doc && docker run --rm -v "$(PWD)/api":/api -w "/api" thethingsindustries/protoc \
	--go_out=. --go_opt=paths=source_relative --go-grpc_out=.  --go-grpc_opt=paths=source_relative \
	--proto_path=. --doc_out=./doc --doc_opt=html,index.html chat.proto

