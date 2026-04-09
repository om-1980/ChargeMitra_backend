package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/om-1980/ChargeMitra_backend/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		service: NewService(db),
	}
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "user not found in token")
		return
	}

	profile, err := h.service.GetProfile(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "user profile fetched successfully", profile)
}