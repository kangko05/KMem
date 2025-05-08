package router

import (
	"gateway/internal/files"

	"github.com/gin-gonic/gin"
)

func setupFileApi(r *gin.Engine, _ *files.UploadQueue) {
	filesAPI := r.Group("/api/files")
	{
		filesAPI.POST("", files.Upload) // upload files
		filesAPI.GET("/:id")            // download files
		filesAPI.GET("")                // list files
	}
}
