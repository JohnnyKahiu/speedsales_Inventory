package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/JohnnyKahiu/speed_sales_proto/auth"
	"google.golang.org/grpc"
)

// InventoryService represents business logic for inventory
type InventoryService struct {
	authClient pb.AuthServiceClient
}

// NewInventoryService creates a new inventory service
func NewInventoryService(authAddr string) (*InventoryService, error) {
	conn, err := grpc.Dial(authAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthServiceClient(conn)

	fmt.Println("client = %v", &client)

	return &InventoryService{
		authClient: client,
	}, nil
}

// ValidateUserToken calls Login.ValidateToken over gRPC
func (s *InventoryService) ValidateUserToken(ctx context.Context, token string) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := s.authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return "", false
	}

	// log.Fatalf("resp = %v,\n", resp)

	if !resp.Valid {
		return "", false
	}

	return resp.Rights, resp.Valid
}
