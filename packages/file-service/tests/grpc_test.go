package tests

import (
	"file-service/internal/server"
	pb "file-service/protogen"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServer(t *testing.T) {
	t.Run("test connection", func(t *testing.T) {
		assert := assert.New(t)

		serv, err := server.New(t.Context())
		assert.Nil(err)
		go serv.Run()

		conn, err := grpc.NewClient("localhost:8001", grpc.WithTransportCredentials(insecure.NewCredentials()))
		assert.Nil(err)
		defer conn.Close()

		client := pb.NewFileServiceClient(conn)

		resp, err := client.Ping(t.Context(), &pb.PingRequest{})
		assert.Nil(err)

		assert.Equal(resp.GetMsg(), "pong")
	})

	// t.Run("test upload", func(t *testing.T) {
	// 	assert := assert.New(t)
	//
	// 	serv, err := server.New(t.Context())
	// 	assert.Nil(err)
	// 	go serv.Run()
	//
	// 	// client side
	// 	conn, err := grpc.NewClient("localhost:8001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// 	assert.Nil(err)
	// 	defer conn.Close()
	//
	// 	client := pb.NewFileServiceClient(conn)
	//
	// 	// build a upload stream
	// 	stream, err := client.Upload(t.Context())
	// 	assert.Nil(err)
	//
	// 	// test data
	// 	testFilename := "test_file.txt"
	// 	testData := [][]byte{
	// 		[]byte("This is the first chunk of test data"),
	// 		[]byte("This is the second chunk of test data"),
	// 		[]byte("This is the final chunk"),
	// 	}
	//
	// 	// send chunks
	// 	for _, chunk := range testData {
	// 		err = stream.Send(&pb.UploadRequest{
	// 			Chunk:    chunk,
	// 			Filename: testFilename,
	// 		})
	// 		assert.Nil(err)
	// 	}
	//
	// 	// close stream and receive data
	// 	reply, err := stream.CloseAndRecv()
	// 	assert.Nil(err)
	// 	assert.Equal(pb.UploadStatus_SUCCESS, reply.GetStatus())
	//
	// 	// get final path from server
	// 	// if its success, should have final path of the file after ":"
	// 	finalPath := strings.TrimSpace(strings.Split(reply.GetMsg(), ":")[1])
	//
	// 	rb, err := os.ReadFile(finalPath)
	// 	assert.Nil(err)
	// 	assert.Equal(rb, bytes.Join(testData, []byte("")))
	//
	// 	time.Sleep(time.Second)
	// })
}
