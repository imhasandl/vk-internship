package server

import (
	"context"

	"github.com/imhasandl/vk-internship/helper"
	pb "github.com/imhasandl/vk-internship/protos"
	"github.com/imhasandl/vk-internship/subpub"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type apiConfig struct {
	pb.UnimplementedSubPubServer
	Port string
	PubSub *subpub.PubSub  
}

func NewServer(port string, pubsub *subpub.PubSub) *apiConfig {
	return &apiConfig{
		Port: port,
		PubSub: pubsub,
	}
}

func (s *apiConfig) Subscribe(req *pb.SubscribeRequest, stream pb.SubPub_SubscribeServer) error {
	msgChan := make(chan string)

	handler := func(msg interface{}) {
		if data, ok := msg.(string); ok {
			msgChan <- data
		}
	}

	subscription, err := s.PubSub.Subscribe(req.Key, handler)
	if err != nil {
		return helper.RespondWithErrorGRPC(context.Background(), codes.InvalidArgument, "invalid argument", err)
	}

	defer subscription.Unsubscribe()

	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data := <-msgChan:
			if err := stream.Send(&pb.Event{Data: data}); err != nil {
				return helper.RespondWithErrorGRPC(ctx, codes.Internal, "Failed to send message", err)
			}
		}
	}
}

func (s *apiConfig) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	err := s.PubSub.Publish(req.Key, req.Data)
	if err != nil {
		return nil, helper.RespondWithErrorGRPC(ctx, codes.InvalidArgument, "invalid argument", err)
	}

	return &emptypb.Empty{}, nil
}