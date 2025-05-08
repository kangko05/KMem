package files

import (
	"context"
	"gateway/internal/utils"
	pb "gateway/protogen"
	"io"
	"log"
	"mime/multipart"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UploadQueue struct {
	uq     *utils.Queue[*multipart.FileHeader]
	conn   *grpc.ClientConn
	client pb.FileServiceClient
}

func NewUploadQueue() (*UploadQueue, error) {
	conn, err := grpc.NewClient(utils.FILESERVICE, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &UploadQueue{
		uq:     utils.NewQueue[*multipart.FileHeader](),
		conn:   conn,
		client: pb.NewFileServiceClient(conn),
	}, nil
}

func (uq *UploadQueue) Add(fh *multipart.FileHeader) {
	uq.uq.Add(fh)
}

// TODO: need better logging for all the return lines
// TODO: Implement structured logging with proper levels (info, error, debug)
// TODO: Add metrics collection for upload attempts, successes, and failures
func (uq *UploadQueue) Run(ctx context.Context) {
	uq.uq.Run(ctx, func(fh *multipart.FileHeader) {
		file, err := fh.Open()
		if err != nil {
			log.Printf("failed to open fileheader: %v", err)
			return
		}
		defer file.Close()

		stream, err := uq.client.Upload(ctx)
		if err != nil {
			log.Printf("failed to connect to file-service: %v", err)
			return
		}

		filename := fh.Filename
		buffer := make([]byte, 1<<20)
		for {
			n, err := file.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("failed to read from file: %v", err)
				return
			}

			if err := stream.Send(&pb.UploadRequest{Chunk: buffer[:n], Filename: filename}); err != nil {
				log.Printf("failed to send stream to file-service: %v", err)
				return
			}
		}

		reply, err := stream.CloseAndRecv()
		if err != nil {
			return
		}

		if reply.GetStatus() == pb.UploadStatus_SUCCESS {
			log.Printf("upload success: %s: %s", reply.GetMsg(), filename)
		} else {
			log.Printf("upload failed: %s: %s", reply.GetMsg(), filename)
			return // err
		}
	})

	go func() {
		select {
		case <-ctx.Done():
			uq.conn.Close()
			log.Println("closing upload queue")
			return
		}
	}()
}
