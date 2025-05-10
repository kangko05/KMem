package main

import (
	"auth-service/dbutils"
	"auth-service/jwtutils"
	pb "auth-service/protogen"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

const PORT = ":8002"

type server struct {
	pb.UnimplementedAuthServiceServer
}

func (s *server) Login(_ context.Context, in *pb.LoginRequest) (*pb.LoginReply, error) {
	userDb, err := dbutils.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer userDb.Close()

	username := in.GetUsername()

	password, exists := userDb.FindUser(username)
	if !exists {
		return &pb.LoginReply{
			Status:  pb.LoginStatus_FAIL,
			Message: fmt.Sprintf("user %s does not exist", username),
		}, fmt.Errorf("user doesn't exist in db")
	}

	if password != dbutils.HashString(in.GetPassword()) {
		return &pb.LoginReply{
			Status:  pb.LoginStatus_FAIL,
			Message: fmt.Sprintf("wrong password"),
		}, fmt.Errorf("wrong password")
	}

	tokenString, err := jwtutils.CreateJWT(username)
	if err != nil {
		return &pb.LoginReply{
			Status:  pb.LoginStatus_FAIL,
			Message: fmt.Sprintf("failed to create access token: %v", err),
		}, fmt.Errorf("wrong password")
	}

	return &pb.LoginReply{
		Status:  pb.LoginStatus_SUCCESS,
		Message: tokenString,
	}, nil
}

func main() {

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0%s", PORT))
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServiceServer(grpcServer, &server{})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
