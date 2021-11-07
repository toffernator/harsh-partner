package main

import (
	"context"
	"errors"
	"flag"
   "fmt"
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
   clients map[string]chan *api.Message
}

func main() {
   grpcServer := grpc.NewServer()
   chatServer := ChatServiceServer{
      Name: somename,
      lamport: lamport.LamportClock{},
      clients: make(map[string](chan *api.Message)),
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
   log.Printf("[%s] Receiving a 'Subscribe' message from %s [%d]", c.Name, in.SubscriberId, c.lamport.Read())

   // Using a goroutine to see if the server is being blocked somehow
   msgs := make(chan *api.Message)
   go notifySubscription(msgs, subscribeServer)

   c.clients[in.SubscriberId] = msgs;

   msg := &api.Message{
      Lamport: &api.Lamport{Time: c.lamport.Read()},
      Content: fmt.Sprintf("%s subscribed! Say hello!", in.SubscriberId),
   }
   c.broadcast(msg)

   // FIXME: Server is being blocked or streams are being closed. Not sure...
   return nil
}

func notifySubscription(msgs chan *api.Message, subscriptionServer api.ChatService_SubscribeServer) {
   for {
      msg := <- msgs
      subscriptionServer.Send(msg)
   }
}

func (c *ChatServiceServer) Unsubscribe(context.Context, *api.UnsubscribeReq) (*api.UnsubscribeResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

func (c *ChatServiceServer) Publish(context.Context, *api.Message) (*api.PublishResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

func (c *ChatServiceServer) broadcast(msg *api.Message) error {
   for clientId, clientChannel := range c.clients {
      c.lamport.Tick()
      log.Printf("[%s] Broadcasting to %s [%d]", c.Name, clientId, c.lamport.Read())
      clientChannel <- msg
   }
   return nil
}
