package url

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	service "github.com/mickeypawis/url-shortener/internal/services/url"
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

func (h *Handler) List(c *gin.Context) {
	urls, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]urlResponse, 0, len(urls))
	for _, u := range urls {
		resp = append(resp, urlResponse{
			ID:          u.ID,
			Code:        u.Code,
			OriginalURL: u.LongURL,
			ShortURL:    h.baseURL + "/" + u.Code,
			CreatedAt:   u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "short url not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
