package app

import (
	pb "github.com/madflojo/tarmac/proto/kvstore"
	"google.golang.org/grpc"
	"net"
)

// Listen will start the GRPC Listener.
func Listen() error {
	l, err := net.Listen("unix", cfg.GetString("grpc_socket_path"))
	if err != nil {
		return err
	}

	srv := grpc.NewServer()
	pb.RegisterKVStoreServer(srv, &GRPCKVServer{})
	err = srv.Serve(l)
	if err != nil {
		return err
	}
	return nil
}
