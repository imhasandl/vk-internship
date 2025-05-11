package main

import (
	"log"
	"net"
	"os"

	pb "github.com/imhasandl/vk-internship/protos"
	"github.com/imhasandl/vk-internship/server"
	"github.com/imhasandl/vk-internship/subpub"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("Set server port in env")
	}

	pubSub := subpub.NewSubPub()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	server := server.NewServer(port, pubSub)
	
	s := grpc.NewServer()
	pb.RegisterSubPubServer(s, server)

	log.Printf("Server listening on %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
