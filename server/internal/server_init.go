package server

import (
	"log"
	"sync"

   "github.com/toffernator/harsh-partner/api"
   "github.com/toffernator/harsh-partner/lamport"
	"google.golang.org/grpc"
)

func NewChatServiceServer(name string) (chatServer *ChatServiceServer) {
   chatServer = &ChatServiceServer{
      Name: name,
      Lamport: lamport.LamportClock{},
      subscribers: map[string]Subscriber{},
      subscriberMutex: sync.Mutex{},
   }

   return chatServer
}

func NewGrpcServer(name string) (gs *grpc.Server, cs *ChatServiceServer) {
      gs = grpc.NewServer();
      cs = NewChatServiceServer(name)

      api.RegisterChatServiceServer(gs, cs)
      log.Println(cs.FmtMsg("Registered"))
      cs.Lamport.Tick()

      return gs, cs
}
