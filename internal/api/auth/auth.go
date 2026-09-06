package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	svcauth "github.com/mickeypawis/url-shortener/internal/services/auth"
)

type Handler struct {
	svc *svcauth.Service
}

func NewHandler(svc *svcauth.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password (min 8 characters) are required"})
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, svcauth.ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, registerResponse{ID: u.ID, Email: u.Email})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, svcauth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{Token: token})
}
