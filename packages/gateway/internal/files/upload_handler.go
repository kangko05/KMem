package files

import (
	"fmt"
	"gateway/internal/utils"
	pb "gateway/protogen"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Upload(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.String(http.StatusBadRequest, "failed to upload: %v", err)
		return
	}
	defer form.RemoveAll()

	files := form.File["files"]
	if len(files) == 0 {
		ctx.String(http.StatusBadRequest, "no files have been uploaded")
		return
	}

	results := make(map[string]string)
	var resultMutex sync.Mutex

	sem := make(chan struct{}, 4) // max 4 concurrent uploads

	var wg sync.WaitGroup

	for _, fileHeader := range files {
		wg.Add(1)
		go handleMultipartFileHeader(ctx, &wg, &resultMutex, &results, sem, fileHeader)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// wait for the result
	select {
	case <-done:
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "completed",
			"results": results,
		})
	case <-time.After(10 * time.Minute): // timeout
		ctx.JSON(http.StatusRequestTimeout, gin.H{
			"status":          "timeout",
			"message":         "Upload operation timed out",
			"partial_results": results,
		})
	}
}

// create client for file-service for each files - with max 4 concurrency - stores resp from file-service into results
// this function was intended to run as goroutine
func handleMultipartFileHeader(ctx *gin.Context, wg *sync.WaitGroup, resultMutex *sync.Mutex, results *map[string]string, sem chan struct{}, fh *multipart.FileHeader) {
	defer wg.Done()

	sem <- struct{}{}
	defer func() { <-sem }()

	conn, err := grpc.NewClient(utils.FILESERVICE, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		resultMutex.Lock()
		(*results)[fh.Filename] = fmt.Sprintf("connection error: %v", err)
		resultMutex.Unlock()
		return
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	file, err := fh.Open()
	if err != nil {
		resultMutex.Lock()
		(*results)[fh.Filename] = fmt.Sprintf("failed to open file: %v", err)
		resultMutex.Unlock()
		return
	}
	defer file.Close()

	stream, err := client.Upload(ctx.Request.Context())
	if err != nil {
		resultMutex.Lock()
		(*results)[fh.Filename] = fmt.Sprintf("failed to create upload stream: %v", err)
		resultMutex.Unlock()
		return
	}

	buffer := make([]byte, 1<<20)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			resultMutex.Lock()
			(*results)[fh.Filename] = fmt.Sprintf("failed to read file: %v", err)
			resultMutex.Unlock()
			return
		}

		if err := stream.Send(&pb.UploadRequest{
			Chunk:    buffer[:n],
			Filename: fh.Filename,
		}); err != nil {
			resultMutex.Lock()
			(*results)[fh.Filename] = fmt.Sprintf("failed to send chunk: %v", err)
			resultMutex.Unlock()
			return
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		resultMutex.Lock()
		(*results)[fh.Filename] = fmt.Sprintf("failed to complete upload: %v", err)
		resultMutex.Unlock()
		return
	}

	resultMutex.Lock()
	if reply.GetStatus() == pb.UploadStatus_SUCCESS {
		(*results)[fh.Filename] = "success: " + reply.GetMsg()
	} else {
		(*results)[fh.Filename] = "failed: " + reply.GetMsg()
	}
	resultMutex.Unlock()
}
