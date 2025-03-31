package database

import (
	"fmt"
	"log"
	"time"

	"github.com/KoLili12/bulb-server/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase создает новое подключение к базе данных
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Настройка логгера GORM
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // Порог для логирования медленных запросов
			LogLevel:      logger.Info, // Уровень логирования
			Colorful:      true,        // Цветной вывод
		},
	)

	// Открытие соединения с БД
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to database successfully")

	// Настройка пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Максимальное количество открытых соединений
	sqlDB.SetMaxOpenConns(25)
	// Максимальное количество простаивающих соединений
	sqlDB.SetMaxIdleConns(10)
	// Максимальное время жизни соединения
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
