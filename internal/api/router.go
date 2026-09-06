package api

import (
	"github.com/gin-gonic/gin"

	apiauth "github.com/mickeypawis/url-shortener/internal/api/auth"
	apiurl "github.com/mickeypawis/url-shortener/internal/api/url"
)

func NewRouter(h *apiurl.Handler, authH *apiauth.Handler) *gin.Engine {
	r := gin.Default()

	r.POST("/api/shorten", h.Shorten)
	r.POST("/api/register", authH.Register)
	r.POST("/api/login", authH.Login)
	r.GET("/:code", h.Redirect)

	return r
}
