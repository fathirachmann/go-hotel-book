package handler

import "github.com/gin-gonic/gin"

// BindRoutes attaches booking endpoints to router.
func (h *Handler) BindRoutes(r *gin.Engine) {
	booking := r.Group("/bookings")
	{
		booking.GET("", h.GetMyBookings)
		booking.POST("", h.PostBooking)
		booking.GET(":id", h.GetBookingDetail)
		booking.POST(":id/checkin", h.PostCheckIn)
		booking.POST(":id/refund", h.PostRefund)
	}
}
