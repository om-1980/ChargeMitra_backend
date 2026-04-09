package stations

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

func (h *Handler) Create(c *gin.Context) {
	var req CreateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	ownerID := c.GetString("user_id")
	if ownerID == "" {
		response.Unauthorized(c, "user not found in token")
		return
	}

	station, err := h.service.Create(ownerID, req)
	if err != nil {
		response.InternalServerError(c, "failed to create station", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "station created successfully", station)
}

func (h *Handler) List(c *gin.Context) {
	stations, err := h.service.List()
	if err != nil {
		response.InternalServerError(c, "failed to fetch stations", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "stations fetched successfully", stations)
}

func (h *Handler) MyStations(c *gin.Context) {
	ownerID := c.GetString("user_id")
	if ownerID == "" {
		response.Unauthorized(c, "user not found in token")
		return
	}

	stations, err := h.service.ListByOwner(ownerID)
	if err != nil {
		response.InternalServerError(c, "failed to fetch owner stations", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "owner stations fetched successfully", stations)
}