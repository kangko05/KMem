package filemanager

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileManager struct {
	tempDir  string
	finalDir string
}

func New() *FileManager {
	fm := &FileManager{
		tempDir:  "/tmp/kmemUpload",
		finalDir: "/tmp/kmemTest", // or whatever
	}

	os.MkdirAll(fm.tempDir, 0755)
	os.MkdirAll(fm.finalDir, 0755)

	return fm
}

// 1. receive uploaded chunks
// 2. write to temp file
// 3. when finished, verify & move to its appropriate dir
func (fm *FileManager) ProcessUpload(filename string) (chan []byte, chan error, chan string) {
	chunkChan := make(chan []byte)
	errChan := make(chan error, 1)
	doneChan := make(chan string, 1) // return final file path when done

	go func() {
		defer close(doneChan)
		defer close(errChan)

		tmpFilePath := filepath.Join(fm.tempDir, filename)
		tmpFile, err := os.Create(tmpFilePath)
		if err != nil {
			errChan <- fmt.Errorf("failed to create temp file: %v", err)
			return
		}
		defer tmpFile.Close()

		var totalSize int64

		for {
			select {
			case chunk, ok := <-chunkChan:
				if !ok {
					// closed chunkChan means all recv process from server is done
					finalFilePath := filepath.Join(fm.finalDir, filename)
					if err := os.Rename(tmpFilePath, finalFilePath); err != nil {
						errChan <- fmt.Errorf("failed to move tmp file to final path: %v", err)
						return
					}

					doneChan <- finalFilePath
					return
				}

				n, err := tmpFile.Write(chunk)
				if err != nil {
					errChan <- fmt.Errorf("failed to write chunk: %v", err)
					return
				}

				totalSize += int64(n)

			case <-errChan:
				tmpFile.Close()
				os.Remove(tmpFilePath)
				return
			}
		}
	}()

	return chunkChan, errChan, doneChan
}
