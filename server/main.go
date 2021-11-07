package main

import (
   "context"
   "errors"
   "chkg.com/chitty-chat/api"
)

type ChatServiceServer struct {
   api.UnimplementedChatServiceServer
}

func (c *ChatServiceServer) Subscribe(*api.SubscribeReq, api.ChatService_SubscribeServer) error {
   return errors.New("Subscribe is not yet implemented")
}

func (c *ChatServiceServer) Unsubscribe(context.Context, *api.UnsubscribeReq) (*api.UnsubscribeResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

func (c *ChatServiceServer) Publish(context.Context, *api.Message) (*api.PublishResp, error) {
   return nil, errors.New("Unsubscribe is not yet implemented")
}

