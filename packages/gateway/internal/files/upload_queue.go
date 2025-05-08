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
	uq     *utils.Queue[*multipart.Form]
	conn   *grpc.ClientConn
	client pb.FileServiceClient
}

func NewUploadQueue() (*UploadQueue, error) {
	conn, err := grpc.NewClient(utils.FILESERVICE, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &UploadQueue{
		uq:     utils.NewQueue[*multipart.Form](),
		conn:   conn,
		client: pb.NewFileServiceClient(conn),
	}, nil
}

func (uq *UploadQueue) Add(fh *multipart.Form) {
	uq.uq.Add(fh)
}

// TODO: need better logging for all the return lines
// TODO: Implement structured logging with proper levels (info, error, debug)
// TODO: Add metrics collection for upload attempts, successes, and failures
func (uq *UploadQueue) Run(ctx context.Context) {
	uq.uq.Run(ctx, func(form *multipart.Form) {
		files := form.File["files"]
		if len(files) == 0 {
			log.Println("no files have been uploaded")
			return
		}
		// defer form.RemoveAll()

		for _, fh := range files {
			file, err := fh.Open()
			if err != nil {
				log.Printf("failed to open fileheader: %v", err)
				return
			}
			// defer file.Close()

			stream, err := uq.client.Upload(ctx)
			if err != nil {
				log.Printf("failed to connect to file-service: %v", err)
				file.Close()
				return
			}

			filename := fh.Filename
			buffer := make([]byte, 1<<20)
			for {
				n, err := file.Read(buffer)
				if err == io.EOF {
					file.Close()
					break
				}
				if err != nil {
					log.Printf("failed to read from file: %v", err)
					file.Close()
					return
				}

				if err := stream.Send(&pb.UploadRequest{Chunk: buffer[:n], Filename: filename}); err != nil {
					log.Printf("failed to send stream to file-service: %v", err)
					file.Close()
					return
				}
			}

			reply, err := stream.CloseAndRecv()
			if err != nil {
				file.Close()
				return
			}

			if reply.GetStatus() == pb.UploadStatus_SUCCESS {
				log.Printf("upload success: %s: %s", reply.GetMsg(), filename)
			} else {
				log.Printf("upload failed: %s: %s", reply.GetMsg(), filename)
				file.Close()
				return // err
			}

			file.Close()
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
