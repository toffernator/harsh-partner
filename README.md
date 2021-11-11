# harsh-partner

## Quick start

You can run:
```
make build-and-start
```
to build a docker image of the server, run it, build the client, and run it.

If you have already built the source files then you can run the program using:
```
make start
```

Whether you start the server in a docker container or not the server starts on
the port `:4042`. The client, by default, attempts to connect to this port.

## Building

To build the client and server run:
```bash
make go-build
```

To build a docker image for the server you can run:
```bash
make docker-build
```

## Running the Program

If you have built the docker image then the server can be run as well as a
client connected to it using:
```bash
make start
```

If you have built the docker image then you can start the server using:
```bash
make server-run
```

If you have built the program from source then the server can be run through
`bin/server` and the client can be run through `bin/client`.

