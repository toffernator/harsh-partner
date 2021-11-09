package server

import (
   "context"
   "errors"
   "log"

   "chkg.com/chitty-chat/api"
)


func (c *ChatServiceServer) Subscribe(in *api.SubscribeReq, stream api.ChatService_SubscribeServer) error {
   c.Lamport.Tick()
   log.Println(c.FmtMsgf("Recieving 'Subscribe' from %s", in.SubscriberId))

   finished := make(chan bool)
   sub := Subscriber{
      Id: in.SubscriberId,
      Stream: stream,
      Finished: finished,
   }
   c.AddSubscriber(sub)
   log.Println(c.FmtMsgf("%s is now subscribed to the server", sub.Id))

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

func (c *ChatServiceServer) Unsubscribe(context.Context, *api.UnsubscribeReq) (*api.UnsubscribeResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

func (c *ChatServiceServer) Publish(context.Context, *api.Message) (*api.PublishResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}
