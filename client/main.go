package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"chkg.com/chitty-chat/api"
	"chkg.com/chitty-chat/lamport"
	"github.com/manifoldco/promptui"
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

   clientLoop(client)
   unsubscribe(client)
}

func clientLoop(c api.ChatServiceClient) {
   isLeaving := false

   for !isLeaving {
      prompt := promptui.Select{
         Label: "Choose an option",
         Items: []string{"Publish", "Leave"},
      }

      _, result, err := prompt.Run()
      if err != nil {
         log.Fatalf("[%s] User Prompt failed with error: %v [%d]", *addressFlag, err, myLamport.Read())
      }

      isLeaving = result == "Leave"
      if !isLeaving {
         msgPrompt := promptui.Prompt{
            Label: "> ",
         }

         msg, err := msgPrompt.Run()
         if err != nil {
            log.Fatalf("[%s] Message Prompt failed with error: %v [%d]", *addressFlag, err, myLamport.Read())
         }

         publish(c, msg)
      }
   }
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
   go listen(messageStream)
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
         log.Printf("[%s] Recieved: %s [%d]", *nameFlag, msg.Content, myLamport.Read())
      }
   }
}

func unsubscribe(c api.ChatServiceClient) {
   myLamport.Tick()
   log.Printf("[%s] Trying to unsubscribe from %s [%d]", *nameFlag, *addressFlag, myLamport.Read())

   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()

   resp, err := c.Unsubscribe(ctx, &api.UnsubscribeReq{
      Lamport: &myLamport.Lamport,
      SubscriberId: *nameFlag,
   })

   if err != nil {
      log.Fatalf("Failed to unsubscribe from %s with error: %v", *addressFlag, err)
   }

   if resp.Status != api.Status_OK {
      log.Fatalf("Could not unsubscribe from %s with server response: %s", *addressFlag, resp.Status)
   }
   myLamport.TickAgainst(resp.Lamport.Time)
   log.Printf("[%s] Unsubscribed. Closing the application [%d]", *nameFlag, myLamport.Read())
   os.Exit(0)
}

func publish(c api.ChatServiceClient, msg string) {
   myLamport.Tick()
   log.Printf("[%s] Trying to publish to %s [%d]", *nameFlag, *addressFlag, myLamport.Read())

   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()

   resp, err := c.Publish(ctx, &api.Message{
      Lamport: &myLamport.Lamport,
      Content: msg,
   })

   if err != nil {
      log.Fatalf("Failed to publish to %s with error: %v", *addressFlag, err)
   }

   if resp.Status != api.Status_OK {
      log.Fatalf("Could not publish to %s with server response: %s", *addressFlag, resp.Status)
   }

   myLamport.TickAgainst(resp.Lamport.Time)
}

func randomName() string {
   somename := "Ben, the Destroyer of Worlds"
   return fmt.Sprintf("%s-%d", somename, time.Now().Unix())
}
