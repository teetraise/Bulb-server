package main

import (
	"log"

	"github.com/KoLili12/bulb-server/internal/database"
	"github.com/KoLili12/bulb-server/internal/handlers"
	"github.com/KoLili12/bulb-server/internal/middleware"
	"github.com/KoLili12/bulb-server/internal/models"
	"github.com/KoLili12/bulb-server/internal/repository"
	"github.com/KoLili12/bulb-server/internal/services"
	"github.com/KoLili12/bulb-server/pkg/config"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Loading configuration...")
	cfg, err := config.LoadConfig("configs")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Устанавливаем режим Gin из конфигурации
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Println("Connecting to database...")
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автоматическая миграция моделей
	log.Println("Running database migrations...")

	log.Println("Starting migration for User...")
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Failed to migrate User: %v", err)
	}

	log.Println("Starting migration for Collection...")
	err = db.AutoMigrate(&models.Collection{})
	if err != nil {
		log.Fatalf("Failed to migrate Collection: %v", err)
	}

	log.Println("Starting migration for Action...")
	err = db.AutoMigrate(&models.Action{})
	if err != nil {
		log.Fatalf("Failed to migrate Action: %v", err)
	}

	log.Println("Initializing Gin...")
	r := gin.Default()

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)

	// Инициализация сервисов
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(cfg)

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(userService, authService)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	log.Println("Setting up routes...")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Маршруты API
	api := r.Group("/api")
	{
		// Открытые маршруты
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Защищенные маршруты
		protected := api.Group("/")
		protected.Use(authMiddleware.RequireAuth())
		{
			// Здесь будут защищенные маршруты
		}
	}

	serverAddr := ":" + cfg.Server.Port
	log.Printf("Starting Bulb API Server on port %s...\n", cfg.Server.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
