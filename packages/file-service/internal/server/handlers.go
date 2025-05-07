package server

import (
	"context"
	pb "file-service/protogen"
	"fmt"
	"io"
)

func (s *server) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	return &pb.PingReply{Msg: "pong"}, nil
}

func (s *server) Upload(stream pb.FileService_UploadServer) error {
	req, err := stream.Recv()
	if err == io.EOF {
		return stream.SendAndClose(&pb.UploadReply{
			Status: pb.UploadStatus_FAIL,
			Msg:    "No data has been received",
		})
	}
	if err != nil {
		return stream.SendAndClose(&pb.UploadReply{
			Status: pb.UploadStatus_FAIL,
			Msg:    err.Error(),
		})
	}

	filename := req.GetFilename()
	firstChunk := req.GetChunk()

	chunkChan, errChan, doneChan := s.fm.ProcessUpload(filename)

	chunkChan <- firstChunk

	go func() {
		defer close(chunkChan)

		for {
			req, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errChan <- err
				return
			}

			chunkChan <- req.GetChunk()
		}
	}()

	select {
	case err := <-errChan:
		return stream.SendAndClose(&pb.UploadReply{
			Status: pb.UploadStatus_FAIL,
			Msg:    err.Error(),
		})
	case finalPath := <-doneChan:
		return stream.SendAndClose(&pb.UploadReply{
			Status: pb.UploadStatus_SUCCESS,
			Msg:    fmt.Sprintf("file upload successful: %s", finalPath),
		})
	}
}
