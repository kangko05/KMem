package files

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
handle upload request - send upload streams to file-service
*/
func Upload(uq *UploadQueue) func(*gin.Context) {
	return func(ctx *gin.Context) {
		form, err := ctx.MultipartForm()
		if err != nil {
			ctx.String(http.StatusBadRequest, "failed to upload: %v", err)
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			ctx.String(http.StatusBadRequest, "no files have been uploaded")
			return
		}

		fileInfo := []gin.H{}

		// send to file-service
		for _, fileHeader := range files {
			log.Printf("filename: %s, file size: %v bytes\n", fileHeader.Filename, fileHeader.Size)

			fileInfo = append(fileInfo, gin.H{
				"filename": fileHeader.Filename,
				"size":     fileHeader.Size,
			})

			uq.Add(fileHeader)
		}

		ctx.JSON(http.StatusOK, "ok")
	}
}
