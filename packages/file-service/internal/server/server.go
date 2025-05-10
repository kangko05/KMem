package server

import (
	"context"
	"file-service/internal/filemanager"
	pb "file-service/protogen"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

// server struct that implements grpc server & methods
type server struct {
	pb.UnimplementedFileServiceServer

	ctx        context.Context
	listener   net.Listener // tcp listener
	grpcServer *grpc.Server
	fm         *filemanager.FileManager
}

func New(ctx context.Context) (*server, error) {
	var serv server

	lis, err := net.Listen("tcp", "0.0.0.0:8001")
	if err != nil {
		return nil, fmt.Errorf("failed to build a tcp listener: %v", err)
	}

	serv.ctx = ctx
	serv.listener = lis
	serv.grpcServer = grpc.NewServer()
	serv.fm = filemanager.New()

	return &serv, nil
}

func (s *server) Run() {
	go func() {
		<-s.ctx.Done()

		log.Println("shutting down the server")
		s.grpcServer.GracefulStop()
		s.listener.Close()

		return
	}()

	pb.RegisterFileServiceServer(s.grpcServer, s)
	s.grpcServer.Serve(s.listener)
}
