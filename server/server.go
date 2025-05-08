package server

import (
	"context"

	pb "github.com/imhasandl/vk-internship/protos"
	"google.golang.org/protobuf/types/known/emptypb"
)

type apiConfig struct {
	pb.UnimplementedSubPubServer
	Port string
}

func NewServer(port string) *apiConfig {
	return &apiConfig{
		Port: port,
	}
}

func (s *apiConfig) Subscribe(req *pb.SubscribeRequest, stream pb.SubPub_SubscribeServer) error {
	return nil
}

func (s *apiConfig) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	return nil, nil
}