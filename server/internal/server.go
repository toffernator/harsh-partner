package server

import (
   "fmt"
	"sync"

	"chkg.com/chitty-chat/api"
   "chkg.com/chitty-chat/lamport"
)

type ChatServiceServer struct {
   Name string
   Lamport lamport.LamportClock
   // Not exported to ensure that it is accessed wrt. Mutex locking
   subscribers map[string]Subscriber
   subscriberMutex sync.Mutex

   api.UnimplementedChatServiceServer
}

// Subscriber pattern is adapted from: github.com/omri86/longlived-grpc
type Subscriber struct {
   Id string
   // Stream is the server side of the RPC stream.
   Stream api.ChatService_SubscribeServer
   // Finished signals the closure of a client's goroutine subscription to a
   // server
   Finished chan<-bool
}

// AddSubscriber adds a new subscriber to the server, respecting concurrent
// resource access.
func (c *ChatServiceServer) AddSubscriber(s Subscriber) {
   c.subscriberMutex.Lock()
   defer c.subscriberMutex.Unlock()

   c.subscribers[s.Id] = s
}

// RemoveSubscriber removes a subscriber from the server, respecting concurrent
// resource access.
func (c *ChatServiceServer) removeSubscriber(id string) {
   c.subscriberMutex.Lock()
   defer c.subscriberMutex.Unlock()

   delete(c.subscribers, id);
}

// broadcast sends a message to any open stream of all subscribes subscribed to
// c.
func (c *ChatServiceServer) Broadcast(msg string) error {
   for _, subscriber := range c.subscribers {
      msg := &api.Message{
         Content: msg,
         Lamport: &api.Lamport{Time: c.Lamport.Read()},
      }

      subscriber.Stream.Send(msg)
      c.Lamport.Tick()
   }
   return nil
}

// FmtMsg returns a String of the standard format [c.Name] msg [c.lamport.time]
func (c *ChatServiceServer) FmtMsg (msg string) string {
   return fmt.Sprintf("[%s] %s [%d]", c.Name, msg, c.Lamport.Read())
}

// FmtMsgf is the Printf equivalent of FmtMsg
func (c *ChatServiceServer) FmtMsgf (format string, v ...interface{}) string {
   intermediaryMsg := c.FmtMsg(format)
   return fmt.Sprintf(intermediaryMsg, v...)
}

