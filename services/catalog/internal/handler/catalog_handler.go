package handler

import (
	"catalog/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CatalogHandler exposes HTTP endpoints for catalog operations.
type CatalogHandler struct {
	svc *service.CatalogService
}

// NewCatalogHandler constructs a CatalogHandler instance.
func NewCatalogHandler(svc *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

// Seed populates baseline catalog data for quick manual testing.
// CreateRoomType adds a new room type with non-UUID auto-increment ID.
func (h *CatalogHandler) CreateRoomType(c *gin.Context) {
	var body struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		BasePrice   int64  `json:"base_price" binding:"required,gt=0"`
		Capacity    int    `json:"capacity" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rt, err := h.svc.CreateRoomType(c.Request.Context(), service.CreateRoomTypeInput{
		Name:        body.Name,
		Description: body.Description,
		BasePrice:   body.BasePrice,
		Capacity:    body.Capacity,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rt})
}

// Availability returns available room types for the requested range.
func (h *CatalogHandler) Availability(c *gin.Context) {
	checkInStr := c.Query("check_in")
	checkOutStr := c.Query("check_out")

	from, errIn := time.Parse("2006-01-02", checkInStr)
	to, errOut := time.Parse("2006-01-02", checkOutStr)
	if errIn != nil || errOut != nil || !to.After(from) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date range"})
		return
	}

	var guests int
	if raw := c.Query("guests"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			guests = v
		}
	}

	items, err := h.svc.Availability(c.Request.Context(), from, to, guests)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}
