package main

import (
	"catalog/internal/entity"
	"catalog/internal/handler"
	"catalog/internal/repo"
	"catalog/internal/service"
	"log"
	"net/http"
	"os"

	"pkg/dbx"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := dbx.InitDatabase("DB_SCHEMA")
	if err != nil {
		log.Fatalf("connect catalog database: %v", err)
	}
	// Drop and recreate tables to safely migrate ID type changes in development
	_ = db.Migrator().DropTable(&entity.RoomInventory{}, &entity.RoomType{})
	if err := db.AutoMigrate(&entity.RoomType{}, &entity.RoomInventory{}); err != nil {
		log.Fatalf("auto migrate catalog schema: %v", err)
	}

	rtRepo := repo.NewRoomTypeRepository(db)
	invRepo := repo.NewInventoryRepository(db)
	svc := service.NewCatalogService(rtRepo, invRepo)
	h := handler.NewCatalogHandler(svc)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	// Replace seeder with room type creation endpoint
	r.POST("/catalog/room-types", h.CreateRoomType)
	r.GET("/catalog/availability", h.Availability)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("catalog service failed: %v", err)
	}
}
