package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/om-1980/ChargeMitra_backend/configs"
	"github.com/om-1980/ChargeMitra_backend/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, cfg *configs.Config) *Handler {
	return &Handler{
		service: NewService(db, cfg.JWTSecret),
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	res, err := h.service.Register(req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, "user registered successfully", res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	res, err := h.service.Login(req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "login successful", res)
}