package main

import (
	"context"
	"errors"
	"flag"
   "fmt"
	"log"
	"net"
   "sync"

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
   Name string
   lamport lamport.LamportClock
   subscribers map[string]broadcastSubscriber
   subscriberMutex sync.Mutex

   api.UnimplementedChatServiceServer
}

// Subscriber pattern is adapted from: github.com/omri86/longlived-grpc
type broadcastSubscriber struct {
   id string
   // stream is the server side of the RPC stream
   stream api.ChatService_SubscribeServer
   // finished is used to signal the closure of a client subscribing goroutine
   finished chan<-bool
}

func main() {
   grpcServer := grpc.NewServer()
   chatServer := ChatServiceServer{
      Name: somename,
      lamport: lamport.LamportClock{},
      subscribers: map[string]broadcastSubscriber{},
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


func (c *ChatServiceServer) Subscribe(in *api.SubscribeReq, stream api.ChatService_SubscribeServer) error {
   c.lamport.Tick()
   log.Printf("[%s] Receiving a 'Subscribe' message from %s [%d]", c.Name, in.SubscriberId, c.lamport.Read())

   fin := make(chan bool)
   subscriber := broadcastSubscriber{
      stream: stream,
      finished: fin,
   }

   c.subscribers[in.SubscriberId] = subscriber
   msg := fmt.Sprintf("%s is now subscribed to the server", in.SubscriberId)
   c.broadcast(msg)

   ctx := stream.Context()
   // Keeps this scope alive because once it is exited the stream is closed
   // Use the finished channel to kill the stream e.g. when a subscriber
   // unsubscribes
   for {
      select {
      case <- fin:
      case <- ctx.Done():
         log.Printf("[%s] Closing subscription for %s [%d]", c.Name, in.SubscriberId, c.lamport.Read())
         return nil
      }
   }
}

func (c *ChatServiceServer) Unsubscribe(context.Context, *api.UnsubscribeReq) (*api.UnsubscribeResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

func (c *ChatServiceServer) Publish(context.Context, *api.Message) (*api.PublishResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

// broadcast will send a Message to the open streams of all subscribers of a
// ChatServiceServer
func (c *ChatServiceServer) broadcast(msg string) error {
   for _, subscriber := range c.subscribers {
      contents := fmt.Sprintf("[%s] %s [%d]", c.Name, msg, c.lamport.Read())
      msgPayload := &api.Message{
         Content: contents,
         Lamport: &api.Lamport{Time: c.lamport.Read()},
      }

      subscriber.stream.Send(msgPayload)
      c.lamport.Tick()
   }
   return nil
}

// addSubscriber adds a new subscriber to the server concurrently
func (c *ChatServiceServer) addSubscriber(subscriber broadcastSubscriber) {
   c.subscriberMutex.Lock()
   defer c.subscriberMutex.Unlock()

   c.subscribers[subscriber.id] = subscriber
}

// removeSubscriber removes a new subscriber from the server concurrently
func (c *ChatServiceServer) removeSubscriber(id string) {
   c.subscriberMutex.Lock()
   defer c.subscriberMutex.Unlock()

   delete(c.subscribers, id);
}
