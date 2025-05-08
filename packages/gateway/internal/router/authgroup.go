package router

import "github.com/gin-gonic/gin"

func setupAuthApi(r *gin.Engine) {
	authAPI := r.Group("/auth")
	{
		authAPI.POST("/login")
		authAPI.POST("/refresh")
	}
}
