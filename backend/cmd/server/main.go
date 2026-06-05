package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/luncher4/vpn-panel/internal/config"
	"github.com/luncher4/vpn-panel/internal/database"
	"github.com/luncher4/vpn-panel/internal/handlers"
	"github.com/luncher4/vpn-panel/internal/middleware"
)

func main() {
	cfg := config.Load()
	db := database.Init(cfg.DBPath)
	database.Migrate(db)

	r := gin.Default()

	// Публичные роуты
	r.POST("/api/auth/login", handlers.Login(db, cfg.JWTSecret))

	// Защищённые роуты
	api := r.Group("/api", middleware.JWTAuth(cfg.JWTSecret))
	{
		api.GET("/dashboard", handlers.Dashboard(db))

		users := api.Group("/users")
		users.GET("", handlers.ListUsers(db))
		users.POST("", handlers.CreateUser(db))
		users.PUT("/:id", handlers.UpdateUser(db))
		users.DELETE("/:id", handlers.DeleteUser(db))
		users.GET("/:id/stats", handlers.UserStats(db))

		inbounds := api.Group("/inbounds")
		inbounds.GET("", handlers.ListInbounds(db))
		inbounds.POST("", handlers.CreateInbound(db))
		inbounds.PUT("/:id", handlers.UpdateInbound(db))
		inbounds.DELETE("/:id", handlers.DeleteInbound(db))

		configs := api.Group("/configs")
		configs.GET("/user/:id", handlers.GenerateConfig(db))
		configs.GET("/user/:id/qr", handlers.GenerateQR(db))
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}
