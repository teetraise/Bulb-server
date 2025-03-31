package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/KoLili12/bulb-server/internal/database"
    "github.com/KoLili12/bulb-server/internal/models"
    "github.com/KoLili12/bulb-server/pkg/config"
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
    err = db.AutoMigrate(&models.User{})
    if err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }

    log.Println("Initializing Gin...")
    r := gin.Default()
    
    log.Println("Setting up routes...")
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    
    serverAddr := ":" + cfg.Server.Port
    log.Printf("Starting Bulb API Server on port %s...\n", cfg.Server.Port)
    if err := r.Run(serverAddr); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}