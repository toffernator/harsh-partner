package main

import (
	"flag"
	"fmt"
	"log"
	"time"

   client "github.com/toffernator/harsh-partner/client/internal"
	"github.com/manifoldco/promptui"
	"google.golang.org/grpc"
)

const (
   address = "localhost:4042"
)

var (
   nameFlag = flag.String("name", randomName(), "Enter the name that you want to use in chitty-chat. It must be unique")
   addressFlag = flag.String("address", address, "Enter the address of the chat server that you want to connect to")
)

func main() {
   flag.Parse();

   log.Printf("[%s] Trying to connect to %s", *nameFlag, *addressFlag)
   conn, err := grpc.Dial(*addressFlag, grpc.WithInsecure(), grpc.WithBlock())
   defer conn.Close()
   if err != nil {
      log.Fatalf("[%s] Unable to connect: %v", *nameFlag, err)
   }
   log.Printf("[%s] Connected to %s", *nameFlag, *addressFlag)

   client := client.NewClient(*nameFlag, *addressFlag, conn)

   client.Subscribe()
   clientLoop(client)
   client.Unsubscribe()
}

// clientLoop is a TUI interaction with the client
func clientLoop(c *client.Client) {
   isLeaving := false

   for !isLeaving {
      prompt := promptui.Select{
         Label: "Choose an option",
         Items: []string{"Publish", "Leave"},
      }

      _, result, err := prompt.Run()
      if err != nil {
         log.Fatalln(c.FmtMsgf("User prompt failed with error %v", err))
      }

      isLeaving = result == "Leave"
      // Only two options, so if not leaving must be publishing
      if !isLeaving {
         msgPrompt := promptui.Prompt{
            Label: "> ",
         }

         msg, err := msgPrompt.Run()
         if err != nil {
            log.Fatalln(c.FmtMsgf("Message Prompt failed with error: %v", err))
         }

         c.Publish(msg)
      }
   }
}

func randomName() string {
   somename := "Ben, the Destroyer of Worlds"
   return fmt.Sprintf("%s-%d", somename, time.Now().Unix())
}
