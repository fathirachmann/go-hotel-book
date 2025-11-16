package handler

import "github.com/gin-gonic/gin"

func (h *AuthHandler) BindRoutes(r *gin.Engine) {
	g := r.Group("/api/v1/auth")
	g.POST("/register", h.HandleRegister)
	g.POST("/login", h.HandleLogin)
}
