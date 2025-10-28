package main

import (
	"booking/internal/entity"
	"booking/internal/handler"
	"booking/internal/repo"
	"booking/internal/service"
	"context"
	"os"
	"pkg/dbx"

	"github.com/gin-gonic/gin"
)

// dummy payment publisher to keep compile
type noopPay struct{}

func (noopPay) RequestPayment(ctx context.Context, bookingID string, amount int64, email string) error {
	return nil
}

func main() {
	db, _ := dbx.InitDatabase("DB_SCHEMA")
	bookingRepo := repo.NewBookingRepository(db)
	invRepo := /* TODO: implement inventory repo */ entity.InventoryRepo(nil)
	pay := noopPay{}
	svc := service.NewService(invRepo, bookingRepo, pay)

	r := gin.Default()
	h := handler.NewHandler(svc)

	r.POST("/bookings", h.PostBooking)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8003"
	}

	r.Run(":" + port)
}
