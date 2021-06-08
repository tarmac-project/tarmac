package main

import (
	"context"
	"encoding/base64"
	"fmt"
	pb "github.com/madflojo/tarmac/proto/kvstore"
	"google.golang.org/grpc"
	"os"
)

//export HTTPHandler
func HTTPHandler() int {
	// Validate input
	d, err := base64.StdEncoding.DecodeString(os.Getenv("HTTP_PAYLOAD"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid Payload")
		return 400
	}

	// Connect to the internal gRPC
	c, err := grpc.Dial("unix://"+os.Getenv("SOCKET_FILE_PATH"), grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to upstream")
		return 500
	}
	client := pb.NewKVStoreClient(c)

	if os.Getenv("HTTP_METHOD") == "POST" || os.Getenv("HTTP_METHOD") == "PUT" {
		// Store data in the KV Store
		r, err := client.Set(context.Background(), &pb.SetRequest{
			Key:  "HTTP_PATH",
			Data: d,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to store within KV")
			return 500
		}

		if r.Code > 0 {
			fmt.Fprintf(os.Stderr, "Unable to store within KV")
			return 500
		}

		fmt.Fprintf(os.Stdout, "Success")
	}
	return 200
}

func main() {}
