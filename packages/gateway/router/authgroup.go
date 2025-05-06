package router

import "github.com/gin-gonic/gin"

func setupAuthApi(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/login")
		auth.POST("/refresh")
	}
}
