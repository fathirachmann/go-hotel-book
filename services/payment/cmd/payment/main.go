package main

import (
	"log"
	"net/http"
	"os"

	"payment/internal/entity"
	"payment/internal/handler"
	"payment/internal/repo"
	"payment/internal/service"

	"pkg/dbx"
	"pkg/jwtx"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := dbx.InitDatabase("DB_SCHEMA")
	if err != nil {
		log.Fatalf("connect payment database: %v", err)
	}
	if err := db.AutoMigrate(&entity.Payment{}, &entity.Refund{}); err != nil {
		log.Fatalf("auto migrate payment schema: %v", err)
	}

	pRepo := repo.NewPaymentRepository(db)
	rRepo := repo.NewRefundRepository(db)
	// Booking client base URL from env (defaults inside ctor if empty)
	bClient := repo.NewBookingHTTPClient(os.Getenv("BOOKING_BASE_URL"))
	svc := service.NewPaymentService(pRepo, rRepo, bClient)
	// JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}
	tm := jwtx.New(jwtSecret, "go-hotel-book")

	h := handler.NewHandler(svc, tm)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	h.BindRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("payment service failed: %v", err)
	}
}
