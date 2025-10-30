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
	"pkg/jwtx"

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
	// Auto-migrate schema (no destructive drops)
	if err := db.AutoMigrate(&entity.Booking{}, &entity.BookingItem{}); err != nil {
		log.Fatalf("auto migrate booking schema: %v", err)
	}
	bookingRepo := repo.NewBookingRepository(db)
	invRepo := repo.NewInventoryHTTPRepo("http://catalog:8002")
	pay := noopPay{}
	svc := service.NewService(invRepo, bookingRepo, pay)

	r := gin.Default()
	// JWT
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	tm := jwtx.New(secret, "booking")
	h := handler.NewHandler(svc, tm)

	r.GET("/health", func(c *gin.Context) {
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
