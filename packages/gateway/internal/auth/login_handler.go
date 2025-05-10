package auth

import (
	pb "gateway/protogen"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type User struct {
	username string
	password string
}

func Login(ctx *gin.Context) {
	var user User
	if err := ctx.Bind(&user); err != nil {
		ctx.String(http.StatusBadRequest, "wrong username or password")
		return
	}

	if len(user.username) == 0 || len(user.password) == 0 {
		ctx.String(http.StatusBadRequest, "wrong username or password")
		return
	}

	// connect to auth-service
	conn, err := grpc.NewClient("localhost:8002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "failed to connect to auth-service: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	reply, err := client.Login(ctx, &pb.LoginRequest{Username: user.username, Password: user.password})
	if err != nil {
		ctx.String(http.StatusInternalServerError, "failed receive reply from auth-service: %v", err)
		return
	}

	if reply.Status == pb.LoginStatus_FAIL {
		ctx.String(http.StatusUnauthorized, "invalid user")
		return
	}

	ctx.String(http.StatusOK, "login successful")
}
