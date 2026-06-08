package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/luncher4/vpn-panel/internal/config"
	"github.com/luncher4/vpn-panel/internal/database"
	"github.com/luncher4/vpn-panel/internal/handlers"
	"github.com/luncher4/vpn-panel/internal/middleware"
)

//go:embed web
var webFS embed.FS

func main() {
	cfg := config.Load()
	db := database.Init(cfg.DBPath)
	database.Migrate(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	staticFS, _ := fs.Sub(webFS, "web")

	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") || strings.HasPrefix(c.Request.URL.Path, "/sub/") {
			c.Next()
			return
		}
		p := c.Request.URL.Path
		if p == "/" || p == "/./" {
			p = "/index.html"
		}
		p = strings.TrimPrefix(p, "/")

		data, err := fs.ReadFile(staticFS, p)
		if err == nil {
			c.Data(200, mimeType(p), data)
			c.Abort()
			return
		}
		// SPA fallback
		indexData, _ := fs.ReadFile(staticFS, "index.html")
		c.Data(200, "text/html; charset=utf-8", indexData)
		c.Abort()
	})

	r.POST("/api/auth/login", handlers.Login(db, cfg.JWTSecret))
	r.GET("/sub/:uuid", handlers.GenerateSubscriptionConfig(db))

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
		inbounds.PATCH("/:id/toggle", handlers.ToggleInbound(db))
		inbounds.DELETE("/:id", handlers.DeleteInbound(db))

		plans := api.Group("/plans")
		plans.GET("", handlers.ListPlans(db))
		plans.POST("", handlers.CreatePlan(db))
		plans.DELETE("/:id", handlers.DeletePlan(db))

		subs := api.Group("/subscriptions")
		subs.GET("", handlers.ListSubscriptions(db))
		subs.POST("", handlers.CreateSubscription(db))
		subs.DELETE(":id", handlers.DeleteSubscription(db))


		profiles := api.Group("/profiles")
		profiles.GET("", handlers.ListProfiles(db))
		profiles.POST("", handlers.CreateProfile(db))
		profiles.GET("/:id", handlers.GetProfile(db))
		profiles.PUT("/:id", handlers.UpdateProfile(db))
		profiles.DELETE("/:id", handlers.DeleteProfile(db))
		nodes := api.Group("/nodes")
		nodes.GET("", handlers.ListNodes(db))
		nodes.POST("", handlers.CreateNode(db))
		nodes.PATCH("/:id/toggle", handlers.ToggleNode(db))
		nodes.DELETE("/:id", handlers.DeleteNode(db))

		api.GET("/templates", handlers.GetTemplates())
		api.POST("/validate-config", handlers.ValidateConfig())

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

func mimeType(name string) string {
	switch {
	case strings.HasSuffix(name, ".js"):
		return "text/javascript; charset=utf-8"
	case strings.HasSuffix(name, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(name, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(name, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(name, ".png"):
		return "image/png"
	default:
		return "text/plain; charset=utf-8"
	}
}
