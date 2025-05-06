package router

import "github.com/gin-gonic/gin"

func setupFileApi(r *gin.Engine) {
	files := r.Group("/api/files")
	{
		files.POST("")    // upload files
		files.GET("/:id") // download files
		files.GET("")     // list files
	}
}
