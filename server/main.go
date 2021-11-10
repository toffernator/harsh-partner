package main

import (
	"flag"
	"log"
	"net"

	"github.com/toffernator/harsh-partner/server/internal"
)

const (
   somename = "Susan"
   address = "localhost:4042"
)

var (
   nameFlag = flag.String("name", somename, "The name by which to recognize the server")
   addressFlag = flag.String("address", address, "The address on which to host the server")
)

func main() {
   grpcServer, chatServer := server.NewGrpcServer(*nameFlag)


   // Listen to requests from the specified address
   lis, err := net.Listen("tcp", address)
   if err != nil {
      log.Fatalf(chatServer.FmtMsgf("Failed to listen on %s with error: %v", address, err))
   }
   defer func() {
      lis.Close()
      log.Printf(chatServer.FmtMsg("Stopped"))
   }()

   log.Printf(chatServer.FmtMsg("Ready to serve requests"))
   if err = grpcServer.Serve(lis); err != nil {
      log.Fatalln(chatServer.FmtMsgf("Failed to server with error: %v", err))
   }
}
