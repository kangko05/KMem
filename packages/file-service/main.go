package main

import (
	"context"
	pb "file-service/protogen"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedFileServiceServer

	ctx         context.Context
	tcpListener *net.Listener
	grpcServer  *grpc.Server
}

func (s *Server) Upload(stream pb.FileService_UploadServer) error {
	var chunks []byte

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UploadReply{
				Status: pb.UploadStatus_SUCCESS,
				Msg:    "File upload success",
			})
		}
		if err != nil {
			return err
		}

		chunks = append(chunks, req.GetChunk()...)
	}
}

func (s *Server) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	return &pb.PingReply{
		Msg: "pong",
	}, nil
}

func NewServer(ctx context.Context) *Server {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost%s", PORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	serv := &Server{
		ctx:         ctx,
		tcpListener: &lis,
		grpcServer:  grpcServer,
	}

	return serv
}

func (s *Server) Run() {
	go func() {
		<-s.ctx.Done()

		fmt.Println("stopping server...")
		s.grpcServer.Stop()
		return
	}()

	pb.RegisterFileServiceServer(s.grpcServer, s)
	s.grpcServer.Serve(*s.tcpListener)
}

const PORT = ":8001"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serv := NewServer(ctx)
	serv.Run()
}
