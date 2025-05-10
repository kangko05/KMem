package router

import (
	"gateway/internal/auth"

	"github.com/gin-gonic/gin"
)

func setupAuthApi(r *gin.Engine) {
	authAPI := r.Group("/auth")
	{
		authAPI.POST("/login", auth.Login)
		authAPI.POST("/refresh")
	}
}
