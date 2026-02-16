package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	pb "github.com/JohnnyKahiu/speed_sales_proto/inventory"
	"github.com/JohnnyKahiu/speedsales_inventory/internal/search"
	"google.golang.org/grpc"
)

// InventoryService implements the InventoryService gRPC interface
type InventoryService struct {
	pb.UnimplementedInventoryServiceServer
}

// SearchProduct searches for products based on a query
func (s *InventoryService) SearchProduct(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	log.Printf("SearchProduct called with query: %s", req.QueryString)

	searchP := search.Search{}

	json.Unmarshal([]byte(req.QueryString), &searchP)

	fmt.Println("\t search product item_code = ", searchP.ItemCode)

	searchP.SearchProduct()

	jstr, err := json.Marshal(searchP.Value)
	if err != nil {
		return &pb.SearchResponse{}, err
	}

	return &pb.SearchResponse{
		Result: string(jstr),
	}, nil
}

// NewServer starts the gRPC server
func NewServer(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterInventoryServiceServer(s, &InventoryService{})

	log.Println("Inventory service running on ", address)
	return s.Serve(lis)
}
