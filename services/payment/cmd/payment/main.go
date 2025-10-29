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

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := dbx.InitDatabase("DB_SCHEMA")
	if err != nil {
		log.Fatalf("connect payment database: %v", err)
	}
	// Dev reset: ensure schema matches new spec by dropping legacy tables
	_ = db.Migrator().DropTable(&entity.Refund{}, &entity.Payment{})
	_ = db.Exec("DROP TABLE IF EXISTS \"payment\".\"refunds\" CASCADE").Error
	_ = db.Exec("DROP TABLE IF EXISTS \"payment\".\"payments\" CASCADE").Error
	if err := db.AutoMigrate(&entity.Payment{}, &entity.Refund{}); err != nil {
		log.Fatalf("auto migrate payment schema: %v", err)
	}

	pRepo := repo.NewPaymentRepository(db)
	rRepo := repo.NewRefundRepository(db)
	// Use internal Docker DNS for booking service without env
	bClient := repo.NewBookingHTTPClient("")
	svc := service.NewPaymentService(pRepo, rRepo, bClient)
	h := handler.NewHandler(svc)

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
