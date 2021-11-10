package server

import (
   "fmt"
   "log"
	"sync"

   "github.com/toffernator/harsh-partner/api"
   "github.com/toffernator/harsh-partner/lamport"
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

   c.Lamport.Tick()
   c.subscribers[s.Id] = s
}

// RemoveSubscriber removes a subscriber from the server, respecting concurrent
// resource access.
func (c *ChatServiceServer) RemoveSubscriber(id string) {
   c.subscriberMutex.Lock()
   defer c.subscriberMutex.Unlock()

   c.Lamport.Tick()
   delete(c.subscribers, id);
}

// broadcast sends a message to any open stream of all subscribes subscribed to
// c.
func (c *ChatServiceServer) Broadcast(msg string) error {
   for _, subscriber := range c.subscribers {
      c.Lamport.Tick()
      msg := &api.Message{
         Content: msg,
         Lamport: &api.Lamport{Time: c.Lamport.Read()},
      }

      log.Println(c.FmtMsgf("Broadcasting to %s", subscriber.Id))
      subscriber.Stream.Send(msg)
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

