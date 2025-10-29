package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"payment/internal/entity"
	"payment/internal/handler"
	"payment/internal/repo"
	"payment/internal/service"

	"pkg/dbx"

	"github.com/gin-gonic/gin"
	"github.com/midtrans/midtrans-go"
)

func main() {
	db, err := dbx.InitDatabase("DB_SCHEMA")
	if err != nil {
		log.Fatalf("connect payment database: %v", err)
	}
	if err := db.AutoMigrate(&entity.Payment{}); err != nil {
		log.Fatalf("auto migrate payment schema: %v", err)
	}

	repository := repo.NewPaymentRepository(db)

	serverKey := strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY"))
	env := strings.ToLower(strings.TrimSpace(os.Getenv("MIDTRANS_ENV")))
	mtEnv := midtrans.Sandbox
	if env == "production" {
		mtEnv = midtrans.Production
	}

	var gateway service.SnapGateway
	if serverKey == "" {
		log.Println("MIDTRANS_SERVER_KEY missing, using mock payment gateway")
		gateway = service.MockSnapClient{}
	} else {
		gateway = service.NewSnapClient(serverKey, mtEnv)
	}

	svc := service.NewPaymentService(repository, gateway, serverKey)
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
