# Runs the protoc command to compile protofiles
proto-compile:
	@mkdir -p api/doc && docker run --rm -v "$(PWD)/api":/api -w "/api" thethingsindustries/protoc \
	--go_out=. --go_opt=paths=source_relative --go-grpc_out=.  --go-grpc_opt=paths=source_relative \
	--proto_path=. --doc_out=./doc --doc_opt=html,index.html chat.proto


# Builds the Client and the Server
go-build:
	@mkdir -p bin &&\
	go build -o "bin/client" client/main.go &&\
	go build -o "bin/server" server/main.go

# Builds the docker image
docker-build:
	@docker build -t harsh-partner-server .

# Run the Server in a docker container
server-run:
	@docker run --rm -dp 4042:4042 harsh-partner-server

# Run the server container interactively.
server-enter:
	@docker run --rm -it -p 4042:4042 harsh-partner-server

# Run the server and start a client connected to the server
build-and-start:
	@make go-build &&\
	make docker-build &&\
	make server-run &&\
	bin/client

# Run the server and start a client connected to the server
start:
	@make server-run && bin/client
