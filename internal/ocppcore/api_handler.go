package ocppcore

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	service *Service
}

func NewAPIHandler(s *Service) *APIHandler {
	return &APIHandler{service: s}
}

func (h *APIHandler) ListEvents(c *gin.Context) {
	filters := map[string]string{
		"ocpp_id":    c.Query("ocpp_id"),
		"direction":  c.Query("direction"),
		"action":     c.Query("action"),
		"message_id": c.Query("message_id"),
		"from":       c.Query("from"),
		"to":         c.Query("to"),
	}

	limit := 20
	offset := 0

	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			if parsed > 200 {
				parsed = 200
			}
			limit = parsed
		}
	}

	if v := c.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	data, err := h.service.ListOCPPEventsAdvanced(filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   data,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *APIHandler) Summary(c *gin.Context) {
	ocppID := c.Param("ocppId")
	if ocppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ocppId is required"})
		return
	}

	data, err := h.service.GetOCPPSummary(ocppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *APIHandler) RemoteStart(c *gin.Context) {
	var req struct {
		OCPPID      string `json:"ocpp_id"`
		IDTag       string `json:"id_tag"`
		ConnectorID *int   `json:"connector_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.OCPPID == "" || req.IDTag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ocpp_id and id_tag are required"})
		return
	}

	messageID, err := h.service.RemoteStartTransaction(req.OCPPID, req.IDTag, req.ConnectorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "RemoteStartTransaction sent",
		"message_id": messageID,
		"ocpp_id":    req.OCPPID,
	})
}

func (h *APIHandler) RemoteStop(c *gin.Context) {
	var req struct {
		OCPPID        string `json:"ocpp_id"`
		TransactionID int64  `json:"transaction_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.OCPPID == "" || req.TransactionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ocpp_id and valid transaction_id are required"})
		return
	}

	messageID, err := h.service.RemoteStopTransaction(req.OCPPID, req.TransactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "RemoteStopTransaction sent",
		"message_id":     messageID,
		"ocpp_id":        req.OCPPID,
		"transaction_id": req.TransactionID,
	})
}