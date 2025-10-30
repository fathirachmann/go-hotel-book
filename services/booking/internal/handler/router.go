package handler

import "github.com/gin-gonic/gin"

// BindRoutes attaches booking endpoints to router.
func (h *Handler) BindRoutes(r *gin.Engine) {
	booking := r.Group("/bookings")
	booking.Use(h.authMiddleware())
	{
		booking.GET("", h.GetMyBookings)
		booking.POST("", h.PostBooking)
		booking.GET("/:id", h.GetBookingDetail)
		booking.DELETE("/:id", h.DeleteBooking)
		booking.POST("/:id/checkin", h.PostCheckIn)
		booking.POST("/:id/checkout", h.PostCheckOut)
		booking.POST("/:id/refund", h.PostRefund)
	}
	internal := r.Group("/internal/bookings")
	{
		internal.POST(":id/status", h.PostInternalUpdateStatus)
	}
}
