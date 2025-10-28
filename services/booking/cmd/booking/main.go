package main

import (
	"booking/internal/entity"
	"booking/internal/handler"
	"booking/internal/repo"
	"booking/internal/service"
	"context"
	"log"
	"net/http"
	"os"

	"pkg/dbx"

	"github.com/gin-gonic/gin"
)

type noopPay struct{}

func (noopPay) RequestPayment(ctx context.Context, bookingID string, amount int64, email string) error {
	return nil
}

func (noopPay) RefundPayment(ctx context.Context, bookingID string, amount int64, reason string) error {
	return nil
}

func main() {
	db, err := dbx.InitDatabase("DB_SCHEMA")
	if err != nil {
		log.Fatalf("connect booking database: %v", err)
	}
	bookingRepo := repo.NewBookingRepository(db)
	invRepo := /* TODO: implement inventory repo */ entity.InventoryRepo(nil)
	pay := noopPay{}
	svc := service.NewService(invRepo, bookingRepo, pay)

	r := gin.Default()
	h := handler.NewHandler(svc)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	h.BindRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("booking service failed: %v", err)
	}
}
