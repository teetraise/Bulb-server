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
	log.Println("🚀 Starting Bulb API Server...")
	
	// Загрузка конфигурации
	log.Println("📋 Loading configuration...")
	cfg, err := config.LoadConfig("configs")
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	log.Printf("✅ Configuration loaded for environment: %s", cfg.Server.Mode)

	// Устанавливаем режим Gin из конфигурации
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("🔧 Gin set to release mode")
	} else {
		log.Println("🔧 Gin set to debug mode")
	}

	// Подключение к базе данных
	log.Println("🗄️  Connecting to database...")
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Автоматическая миграция моделей
	log.Println("🔄 Running database migrations...")

	log.Println("  📝 Migrating User model...")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("❌ Failed to migrate User: %v", err)
	}

	log.Println("  📝 Migrating Collection model...")
	if err := db.AutoMigrate(&models.Collection{}); err != nil {
		log.Fatalf("❌ Failed to migrate Collection: %v", err)
	}

	log.Println("  📝 Migrating Action model...")
	if err := db.AutoMigrate(&models.Action{}); err != nil {
		log.Fatalf("❌ Failed to migrate Action: %v", err)
	}

	log.Println("✅ Database migrations completed successfully")

	// Инициализация Gin
	log.Println("🌐 Initializing Gin router...")
	r := gin.Default()

	// CORS middleware
	log.Println("🔧 Setting up CORS middleware...")
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Инициализация репозиториев
	log.Println("🗃️  Initializing repositories...")
	userRepo := repository.NewUserRepository(db)
	collectionRepo := repository.NewCollectionRepository(db)
	actionRepo := repository.NewActionRepository(db)

	// Инициализация сервисов
	log.Println("⚙️  Initializing services...")
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(cfg)
	collectionService := services.NewCollectionService(collectionRepo, actionRepo, userRepo)

	// Инициализация обработчиков
	log.Println("🎯 Initializing handlers...")
	authHandler := handlers.NewAuthHandler(userService, authService)
	userHandler := handlers.NewUserHandler(userService)
	collectionHandler := handlers.NewCollectionHandler(collectionService)

	// Middleware
	log.Println("🔐 Setting up authentication middleware...")
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Настройка маршрутов
	log.Println("🛣️  Setting up routes...")
	
	// Health check endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "healthy",
			"service": "bulb-api",
		})
	})

	// API маршруты
	api := r.Group("/api")
	{
		// ===== ОТКРЫТЫЕ МАРШРУТЫ (без аутентификации) =====
		
		// Аутентификация
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Публичные коллекции (просмотр без авторизации)
		collections := api.Group("/collections")
		{
			collections.GET("", collectionHandler.List)                    // Список всех коллекций
			collections.GET("/trending", collectionHandler.GetTrending)    // Популярные коллекции
			collections.GET("/:id", collectionHandler.GetByID)             // Коллекция по ID
			collections.GET("/:id/actions", collectionHandler.GetActions)  // Карточки коллекции
			collections.GET("/:id/stats", collectionHandler.GetCollectionStats) // Статистика коллекции
		}

		// Публичная информация о пользователях
		users := api.Group("/users")
		{
			users.GET("/:id", userHandler.GetUserByID) // Публичная информация о пользователе
		}

		// ===== ЗАЩИЩЕННЫЕ МАРШРУТЫ (требуют аутентификации) =====
		
		protected := api.Group("/")
		protected.Use(authMiddleware.RequireAuth())
		{
			// Профиль пользователя
			protected.GET("/me", userHandler.GetProfile)                    // Текущий пользователь
			protected.GET("/user/profile", userHandler.GetProfile)          // Альтернативный эндпоинт
			protected.PUT("/user/profile", userHandler.UpdateProfile)       // Обновление профиля

			// Коллекции пользователя
			protected.GET("/user/collections", collectionHandler.GetUserCollections) // Коллекции пользователя

			// Управление коллекциями
			protected.POST("/collections", collectionHandler.Create)                        // Создание простой коллекции
			protected.POST("/collections/with-actions", collectionHandler.CreateWithActions) // Создание коллекции с карточками
			protected.PUT("/collections/:id", collectionHandler.Update)                     // Обновление коллекции
			protected.DELETE("/collections/:id", collectionHandler.Delete)                  // Удаление коллекции

			// Управление карточками
			protected.POST("/collections/:id/actions", collectionHandler.AddAction)  // Добавление карточки
			protected.DELETE("/actions/:id", collectionHandler.RemoveAction)         // Удаление карточки
		}
	}

	// Логирование всех зарегистрированных маршрутов
	log.Println("📋 Registered routes:")
	log.Println("  🏥 Health: GET /ping")
	log.Println("  🔐 Auth:")
	log.Println("    POST /api/auth/register")
	log.Println("    POST /api/auth/login") 
	log.Println("    POST /api/auth/refresh")
	log.Println("  👥 Users:")
	log.Println("    GET  /api/users/:id")
	log.Println("    GET  /api/me (protected)")
	log.Println("    GET  /api/user/profile (protected)")
	log.Println("    PUT  /api/user/profile (protected)")
	log.Println("  📚 Collections:")
	log.Println("    GET  /api/collections")
	log.Println("    GET  /api/collections/trending")
	log.Println("    GET  /api/collections/:id")
	log.Println("    GET  /api/collections/:id/actions")
	log.Println("    GET  /api/collections/:id/stats")
	log.Println("    GET  /api/user/collections (protected)")
	log.Println("    POST /api/collections (protected)")
	log.Println("    POST /api/collections/with-actions (protected)")
	log.Println("    PUT  /api/collections/:id (protected)")
	log.Println("    DELETE /api/collections/:id (protected)")
	log.Println("  🃏 Actions:")
	log.Println("    POST /api/collections/:id/actions (protected)")
	log.Println("    DELETE /api/actions/:id (protected)")

	// Запуск сервера
	serverAddr := ":" + cfg.Server.Port
	log.Printf("🎉 Starting Bulb API Server on port %s", cfg.Server.Port)
	log.Printf("🌍 API endpoints available at:")
	log.Printf("   • http://localhost:%s/ping (health check)", cfg.Server.Port)
	log.Printf("   • http://localhost:%s/api (main API)", cfg.Server.Port)
	log.Printf("   • Docs: Check the handlers for detailed endpoint documentation")
	
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}