package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"

	"chkg.com/chitty-chat/api"
	"chkg.com/chitty-chat/lamport"
   "google.golang.org/grpc"
)

const (
   somename = "Susan"
   address = "localhost:4042"
   port = "4042"
)

var (
   nameFlag = flag.String("name", somename, "The name by which to recognize the server")
   addressFlag = flag.String("address", address, "The address on which to host the server")
)

type ChatServiceServer struct {
   api.UnimplementedChatServiceServer
   Name string
   lamport lamport.LamportClock
   clients map[string]api.ChatService_SubscribeServer
}

func main() {
   grpcServer := grpc.NewServer()
   chatServer := ChatServiceServer{
      Name: somename,
      lamport: lamport.LamportClock{},
      clients: map[string]api.ChatService_SubscribeServer{},
   }

   api.RegisterChatServiceServer(grpcServer, &chatServer)
   log.Printf("[%s] Registered [%d]", chatServer.Name, chatServer.lamport.Read())
   chatServer.lamport.Tick()

   lis, err := net.Listen("tcp", *addressFlag)
   if err != nil {
      log.Fatalf("[%s] Failed to listen on %s: %v", chatServer.Name, port, err)
   }
   defer func() {
      lis.Close()
      log.Printf("[%s] Stopped", chatServer.Name)
   }()

   log.Printf("[%s] Server Started! [%d]", chatServer.Name, chatServer.lamport.Read())
   if err = grpcServer.Serve(lis); err != nil {
      log.Fatalf("[%s] Failed to serve: %v", chatServer.Name, err)
   }
}

func (c *ChatServiceServer) Subscribe(in *api.SubscribeReq, subscribeServer api.ChatService_SubscribeServer) error {
   c.lamport.Tick()
   log.Printf("[%s] Recieving a 'Subscribe' message from %s [%d]", c.Name, in.SubscriberId, c.lamport.Read())

   c.clients[in.SubscriberId] = subscribeServer;

   return nil
}

func (c *ChatServiceServer) broadcast(msg *api.Message) error {
   for _, client := range c.clients {
      if err := client.Send(msg); err != nil {
         return err;
      }
   }
   return nil
}

func (c *ChatServiceServer) Unsubscribe(context.Context, *api.UnsubscribeReq) (*api.UnsubscribeResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

func (c *ChatServiceServer) Publish(context.Context, *api.Message) (*api.PublishResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

