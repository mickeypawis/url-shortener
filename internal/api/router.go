package api

import (
	"github.com/gin-gonic/gin"

	apiauth "github.com/mickeypawis/url-shortener/internal/api/auth"
	apiurl "github.com/mickeypawis/url-shortener/internal/api/url"
)

func NewRouter(h *apiurl.Handler, authH *apiauth.Handler, authMiddleware gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	r.POST("/api/shorten", h.Shorten)
	r.POST("/api/register", authH.Register)
	r.POST("/api/login", authH.Login)
	r.GET("/:code", h.Redirect)

	protected := r.Group("/api")
	protected.Use(authMiddleware)
	protected.GET("/urls", h.List)
	protected.DELETE("/urls/:id", h.Delete)

	return r
}
