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
func (h *CatalogHandler) Seed(c *gin.Context) {
	if err := h.svc.SeedSample(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
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
