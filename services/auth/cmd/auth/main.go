package main

import (
	"log"
	"net/http"
	"os"

	authhttp "auth/internal/http"
	"auth/internal/repo"
	"auth/internal/usecase"
	"pkg/dbx"
	"pkg/jwtx"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	db, err := dbx.InitDatabase("DB_SCHEMA")
	if err != nil {
		log.Fatalf("init database: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET env is required")
	}

	tokenManager := jwtx.New(jwtSecret, "auth-service")
	userRepo := repo.NewUserRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo, tokenManager)
	handler := authhttp.NewAuthHandler(authUsecase)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "auth"})
	})
	handler.UserRoutes(r)

	log.Printf("auth service listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
