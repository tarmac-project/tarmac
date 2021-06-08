package app

import (
	"context"
	"fmt"
	pb "github.com/madflojo/tarmac/proto/kvstore"
)

// GRPCKVServer is a server which holds GRPC handlers for KV Store.
type GRPCKVServer struct {
	pb.UnimplementedKVStoreServer
}

// Define common errors to return.
var (
	errKeyNotDefined    = fmt.Errorf("Key not defined")
	errFailedFetchData  = fmt.Errorf("Failed to fetch data")
	errFailedStoreData  = fmt.Errorf("Failed to store data")
	errFailedDeleteData = fmt.Errorf("Failed to delete data")
)

// Get will retrieve requested information from the datastore and return it.
func (s *GRPCKVServer) Get(ctx context.Context, msg *pb.GetRequest) (*pb.Data, error) {
	// Create reply message
	r := &pb.Data{
		Status: &pb.Status{
			Code:        0,
			Description: "Success",
		},
	}

	// Check key length
	if len(msg.Key) == 0 {
		r.Status.Code = 4
		r.Status.Description = fmt.Sprintf("%s", errKeyNotDefined)
		return r, nil
	}

	// Fetch data using key
	d, err := kv.Get(msg.Key)
	if err != nil {
		r.Status.Code = 5
		r.Status.Description = fmt.Sprintf("%s", errFailedFetchData)
		return r, nil
	}

	// Return data to client
	r.Data = d
	return r, nil
}

// Set will take the supplied data and store it within the datastore returning success or failure.
func (s *GRPCKVServer) Set(ctx context.Context, msg *pb.SetRequest) (*pb.Status, error) {
	// Create reply message
	r := &pb.Status{
		Code:        0,
		Description: "Success",
	}

	// Check key length
	if len(msg.Key) == 0 {
		r.Code = 4
		r.Description = fmt.Sprintf("%s", errKeyNotDefined)
		return r, nil
	}

	// Insert data into kv store
	err := kv.Set(msg.Key, msg.Data)
	if err != nil {
		r.Code = 5
		r.Description = fmt.Sprintf("%s", errFailedStoreData)
		return r, nil
	}

	return r, nil
}

// Delete will remove the specified key from the datastore and return success or failure.
func (s *GRPCKVServer) Delete(ctx context.Context, msg *pb.DeleteRequest) (*pb.Status, error) {
	// Create reply message
	r := &pb.Status{
		Code:        0,
		Description: "Success",
	}

	// Check key length
	if len(msg.Key) == 0 {
		r.Code = 4
		r.Description = fmt.Sprintf("%s", errKeyNotDefined)
		return r, nil
	}

	err := kv.Delete(msg.Key)
	if err != nil {
		r.Code = 5
		r.Description = fmt.Sprintf("%s", errFailedDeleteData)
	}

	return r, nil
}
