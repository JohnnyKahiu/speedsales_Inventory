package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/JohnnyKahiu/speed_sales_proto/user"
	"google.golang.org/grpc"
)

// LoginService represents business logic for inventory
type LoginService struct {
	authClient pb.AuthServiceClient
}

// NewLoginService creates a new inventory service
func NewLoginService(authAddr string) (*LoginService, error) {
	conn, err := grpc.Dial(authAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthServiceClient(conn)

	fmt.Println("client = %v", &client)

	return &LoginService{
		authClient: client,
	}, nil
}

// ValidateUserToken calls Login.ValidateToken over gRPC
func (s *LoginService) ValidateUserToken(ctx context.Context, token string) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	resp, err := s.authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return "", false
	}

	// log.Println("resp = %v,\n", resp)

	if !resp.Valid {
		return "", false
	}

	return resp.Rights, resp.Valid
}

// FetchUser calls Login.ValidateToken over gRPC
/*func (s *LoginService) FetchUser(ctx context.Context, username string) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := s.userClient.FetchUser(ctx, &pbu.UserRequest{"username": username})
	if err != nil {
		return "", false
	}

	// log.Fatalf("resp = %v,\n", resp)

	if !resp.Valid {
		return "", false
	}

	return resp.Rights, resp.Valid
}*/
