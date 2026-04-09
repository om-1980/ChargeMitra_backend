package chargers

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
	var req CreateChargerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	charger, err := h.service.Create(req)
	if err != nil {
		if err.Error() == "station not found" || err.Error() == "charger with this ocpp_id already exists" {
			response.BadRequest(c, err.Error(), nil)
			return
		}

		response.InternalServerError(c, "failed to create charger", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "charger created successfully", charger)
}

func (h *Handler) List(c *gin.Context) {
	chargers, err := h.service.List()
	if err != nil {
		response.InternalServerError(c, "failed to fetch chargers", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "chargers fetched successfully", chargers)
}

func (h *Handler) ListByStation(c *gin.Context) {
	stationID := c.Param("stationId")
	if stationID == "" {
		response.BadRequest(c, "stationId is required", nil)
		return
	}

	chargers, err := h.service.ListByStation(stationID)
	if err != nil {
		response.InternalServerError(c, "failed to fetch station chargers", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "station chargers fetched successfully", chargers)
}