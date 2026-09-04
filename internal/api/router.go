package api

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()

	r.POST("/api/shorten", h.Shorten)
	r.GET("/:code", h.Redirect)

	return r
}
