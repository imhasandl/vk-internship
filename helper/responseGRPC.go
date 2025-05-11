package helper

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RespondWithErrorGRPC(ctx context.Context, code codes.Code, msg string, err error) error {
	if err != nil {
		log.Println(err)
	}

	if code > codes.Internal {
		log.Printf("Responding with 5XX gRPC error: %s", msg)
	}

	type errorResponse struct {
		ServiceError string `json:"error"`
	}

	jsonBytes, err := json.Marshal(errorResponse{ServiceError: msg})
	if err != nil {
		log.Printf("Error marshalling error JSON: %s", err)
		return status.Errorf(codes.Internal, "Failed to marshal error response")
	}

	log.Printf("ServiceError: %s, Code: %s", string(jsonBytes), code.String()) // Log the error
	return status.Error(code, msg)
}
