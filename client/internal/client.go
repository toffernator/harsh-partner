package client

import (
   "context"
   "fmt"
   "io"
   "log"
   "os"
   "time"

   "chkg.com/chitty-chat/api"
   "chkg.com/chitty-chat/lamport"
	"google.golang.org/grpc"
)

type Client struct {
   Name string
   ConnectedTo string
   Lamport lamport.LamportClock

   chat api.ChatServiceClient
}

func NewClient(name string, connectedTo string, conn *grpc.ClientConn) *Client {
   return &Client{
      Name: name,
      ConnectedTo: connectedTo,
      chat: api.NewChatServiceClient(conn),
   }
}

// subscribe subscribes the client to given address.
func (c *Client) Subscribe() {
   c.Lamport.Tick();
   log.Println(c.FmtMsgf("Trying to subscribe to %s", c.ConnectedTo))
   ctx := context.Background()

   subscribeReq := &api.SubscribeReq{
      SubscriberId: c.Name,
      Lamport: &api.Lamport{Time: c.Lamport.Read()},
   }

   messageStream, err := c.chat.Subscribe(ctx, subscribeReq)
   if err != nil {
      log.Println(c.FmtMsgf("Could not subscribe to: %v", c.Name, err))
   }

   log.Println(c.FmtMsgf("Subscribed to %s", c.Name))
   go c.Listen(messageStream)
}

// listen recieves messages sent into the stream by a server on the other end.
// Listen is blocking.
func (c *Client) Listen(stream api.ChatService_SubscribeClient) {
   for {
      msg, err := stream.Recv()
      if err == io.EOF {
         // Wait before polling the stream again
         time.Sleep(time.Second)
      } else if err != nil {
         log.Fatalln(c.FmtMsgf("Failed to read incoming messages with error: %v", err))
      } else if &msg != nil {
         c.Lamport.TickAgainst(msg.Lamport.GetTime())
         log.Println(c.FmtMsgf("Recieved: %s", msg.Content))
      }
   }
}

// unsubscribe unsubscribes the client from the server that the client is
// connected to.
func (c *Client) Unsubscribe() {
   c.Lamport.Tick()
   log.Println(c.FmtMsgf("Trying to unsubscribe from %s", c.ConnectedTo))

   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()

   resp, err := c.chat.Unsubscribe(ctx, &api.UnsubscribeReq{
      Lamport: &c.Lamport.Lamport,
      SubscriberId: c.Name,
   })

   if err != nil {
      log.Fatalln(c.FmtMsgf("Failed to unsubscribe from %s with error: %v", c.ConnectedTo, err))
   }

   if resp.Status != api.Status_OK {
      log.Fatalln(c.FmtMsgf("Could not unsubscribe from %s with server response: %s", c.ConnectedTo, resp.Status))

   }

   c.Lamport.TickAgainst(resp.Lamport.Time)
   log.Println(c.FmtMsg("Unsubscribed. Closing the application"))
   os.Exit(0)
}

// publish publishes a message to the server to be broadcasted by the server
func (c *Client) Publish(msg string) {
   c.Lamport.Tick()
   log.Println(c.FmtMsgf("Trying to publish to %s", c.ConnectedTo))

   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()

   resp, err := c.chat.Publish(ctx, &api.Message{
      Lamport: &c.Lamport.Lamport,
      Content: msg,
   })

   if err != nil {
      log.Fatalln(c.FmtMsgf("Failed to publish to %s with error: %v", c.ConnectedTo, err))
   }

   if resp.Status != api.Status_OK {
      log.Fatalln(c.FmtMsgf("Could not publish to %s with server response: %s", c.ConnectedTo, resp.Status))
   }

   c.Lamport.TickAgainst(resp.Lamport.Time)
}

func (c *Client) FmtMsg(msg string) string {
   return fmt.Sprintf("[%s] %s [%d]", c.Name, msg, c.Lamport.Read())
}

func (c *Client) FmtMsgf(msg string, v ...interface{}) string {
   msgIntermediary := c.FmtMsg(msg)
   return fmt.Sprintf(msgIntermediary, v)
}
