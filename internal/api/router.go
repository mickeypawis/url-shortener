package api

import (
	"github.com/gin-gonic/gin"

	apiauth "github.com/mickeypawis/url-shortener/internal/api/auth"
	apiurl "github.com/mickeypawis/url-shortener/internal/api/url"
)

func NewRouter(h *apiurl.Handler, authH *apiauth.Handler, authMiddleware gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	r.GET("/:code", h.Redirect)

	api := r.Group("/api")
	api.POST("/shorten", h.Shorten)
	api.POST("/register", authH.Register)
	api.POST("/login", authH.Login)

	protected := api.Group("")
	protected.Use(authMiddleware)
	protected.GET("/urls", h.List)
	protected.DELETE("/urls/:id", h.Delete)

	return r
}
