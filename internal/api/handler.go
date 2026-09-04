package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mickeypawis/url-shortener/internal/service"
)

type Handler struct {
	svc     *service.URLService
	baseURL string
}

func NewHandler(svc *service.URLService, baseURL string) *Handler {
	return &Handler{svc: svc, baseURL: baseURL}
}

func (h *Handler) Shorten(c *gin.Context) {
	var req shortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "original_url is required"})
		return
	}

	u, err := h.svc.Shorten(c.Request.Context(), req.OriginalUrl)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "original_url must be an absolute http(s) URL"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, shortenResponse{
		Code:     u.Code,
		ShortURL: h.baseURL + "/" + u.Code,
	})
}

func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")

	u, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "short url not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Redirect(http.StatusFound, u.LongURL)
}
