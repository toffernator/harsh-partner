package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
   "time"

	"chkg.com/chitty-chat/api"
	"chkg.com/chitty-chat/lamport"
	"google.golang.org/grpc"
)

const (
   address = "localhost:4042"
)

var (
   nameFlag = flag.String("name", randomName(), "Enter the name that you want to use in chitty-chat. It must be unique")
   addressFlag = flag.String("address", address, "Enter the address of the chat server that you want to connect to")
   myLamport = lamport.LamportClock{}
)

func main() {
   flag.Parse();

   log.Printf("[%s] Trying to connect to %s", *nameFlag, *addressFlag)
   conn, err := grpc.Dial(*addressFlag, grpc.WithInsecure(), grpc.WithBlock())
   defer conn.Close()
   if err != nil {
      log.Fatalf("[%s] Unable to connect: %v", *nameFlag, err)
   }
   log.Printf("[%s] Connected to %s [%d]", *nameFlag, *addressFlag, myLamport.Read())

   client := api.NewChatServiceClient(conn)
   subscribe(client)
}

func subscribe(client api.ChatServiceClient) {
   myLamport.Tick();
   log.Printf("[%s] Trying to subscribe to %s [%d]", *nameFlag, *addressFlag, myLamport.Read())
   ctx := context.Background()

   subscribeReq := &api.SubscribeReq{
      SubscriberId: *nameFlag,
      Lamport: &api.Lamport{Time: myLamport.Read()},
   }

   messageStream, err := client.Subscribe(ctx, subscribeReq)
   if err != nil {
      log.Fatalf("[%s] Could not subscribe: %v", *nameFlag, err)
   }

   log.Printf("[%s] Subscribed to %s [%d]", *nameFlag, *addressFlag, myLamport.Read())
   listen(messageStream)
}

func listen(stream api.ChatService_SubscribeClient) {
   for {
      msg, err := stream.Recv()
      if err == io.EOF {
         // Wait before polling the stream again
         time.Sleep(time.Second)
      } else if err != nil {
         log.Fatalf("[%s] Failed to read incoming messages: %v", *nameFlag, err)
      } else if &msg != nil {
         myLamport.TickAgainst(msg.Lamport.GetTime())
         log.Printf("[%s] Recieved: %s [%d]", *nameFlag, msg.Content, msg.Lamport.Time)
      }
   }
}

// Unsubscribe

// Publish

func randomName() string {
   somename := "Ben, the Destroyer of Worlds"
   return fmt.Sprintf("%s-%d", somename, time.Now().Unix())
}
