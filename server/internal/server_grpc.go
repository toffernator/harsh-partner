package server

import (
	"context"
	"fmt"
	"log"

	"chkg.com/chitty-chat/api"
)


func (c *ChatServiceServer) Subscribe(in *api.SubscribeReq, stream api.ChatService_SubscribeServer) error {
   c.Lamport.TickAgainst(in.Lamport.Time)
   log.Println(c.FmtMsgf("Recieving 'Subscribe' from %s", in.SubscriberId))

   finished := make(chan bool)
   sub := Subscriber{
      Id: in.SubscriberId,
      Stream: stream,
      Finished: finished,
   }
   c.AddSubscriber(sub)
   log.Println(c.FmtMsgf("%s is now subscribed to the server", sub.Id))
   c.Broadcast(fmt.Sprintf("%s has joined, say hello!", sub.Id))

   // Keeps this scope alive because once it is exited the stream is closed
   // Use the finished channel to kill the stream e.g. when a subscriber
   // unsubscribes
   ctx := stream.Context()
   for {
      select {
      case <- finished:
      case <- ctx.Done():
         log.Println(c.FmtMsgf("Closing the subscription for %s", sub.Id))
         return nil
      }
   }
}

func (c *ChatServiceServer) Unsubscribe(ctx context.Context, in *api.UnsubscribeReq) (*api.UnsubscribeResp, error) {
   c.Lamport.TickAgainst(in.Lamport.Time)
   log.Println(c.FmtMsgf("Recieving 'Unsubscribe' from %s", in.SubscriberId))

   c.RemoveSubscriber(in.SubscriberId)
   log.Println(c.FmtMsgf("%s has unsubscribed from the server", in.SubscriberId))
   c.Broadcast(fmt.Sprintf("%s has joined, say hello!", in.SubscriberId))

   return &api.UnsubscribeResp{
      Lamport: &c.Lamport.Lamport,
      Status: api.Status_OK,
   }, nil
}

func (c *ChatServiceServer) Publish(ctx context.Context, msg *api.Message) (*api.PublishResp, error) {
   c.Lamport.TickAgainst(msg.Lamport.Time)
   log.Println(c.FmtMsg("Recieving 'Publish'"))

   c.Broadcast(msg.Content)

   return &api.PublishResp{
      Lamport: &c.Lamport.Lamport,
      Status: api.Status_OK,
   }, nil
}
