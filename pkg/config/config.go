package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// Config содержит все конфигурационные параметры приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig содержит настройки HTTP-сервера
type ServerConfig struct {
	Port string
	Mode string
}

// DatabaseConfig содержит настройки подключения к базе данных
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig содержит настройки для JWT-аутентификации
type JWTConfig struct {
	Secret    string
	ExpiresIn int // время жизни токена в часах
}

// LoadConfig загружает конфигурацию из файла config.yml в указанной директории
func LoadConfig(path string) (*Config, error) {
	// Определяем среду выполнения
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	log.Printf("Loading configuration for environment: %s", env)

	// Если production или есть переменные базы данных - используем переменные окружения
	if env == "production" || os.Getenv("DATABASE_HOST") != "" {
		log.Println("Using environment variables for configuration")
		return loadFromEnv()
	}

	log.Println("Using config file for configuration")
	return loadFromFile(path)
}

// loadFromFile загружает конфигурацию из файла (для development)
func loadFromFile(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	// Установка значений по умолчанию
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("jwt.expiresin", 24)

	// Чтение файла конфигурации
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Преобразование конфигурации в структуру
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	log.Println("Configuration loaded successfully from file")
	return config, nil
}

// loadFromEnv загружает конфигурацию только из переменных окружения
func loadFromEnv() (*Config, error) {
	// Получаем порт (Railway может переопределить через PORT)
	port := getEnvOrDefault("PORT", "8080")

	// Получаем JWT ExpiresIn
	expiresIn := 24
	if expiresInStr := os.Getenv("JWT_EXPIRES_IN"); expiresInStr != "" {
		if parsed, err := strconv.Atoi(expiresInStr); err == nil {
			expiresIn = parsed
		}
	}

	config := &Config{
		Server: ServerConfig{
			Port: port,
			Mode: "release",
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DATABASE_HOST"),
			Port:     getEnvOrDefault("DATABASE_PORT", "5432"),
			User:     os.Getenv("DATABASE_USER"),
			Password: os.Getenv("DATABASE_PASSWORD"),
			Name:     os.Getenv("DATABASE_NAME"),
			SSLMode:  getEnvOrDefault("DATABASE_SSL_MODE", "require"),
		},
		JWT: JWTConfig{
			Secret:    os.Getenv("JWT_SECRET"),
			ExpiresIn: expiresIn,
		},
	}

	// Логируем что получили (без паролей)
	log.Printf("Database Host: %s", config.Database.Host)
	log.Printf("Database Port: %s", config.Database.Port)
	log.Printf("Database User: %s", config.Database.User)
	log.Printf("Database Name: %s", config.Database.Name)
	log.Printf("Server Port: %s", config.Server.Port)

	// Проверяем обязательные поля
	if config.Database.Host == "" {
		return nil, fmt.Errorf("DATABASE_HOST environment variable is required")
	}
	if config.Database.User == "" {
		return nil, fmt.Errorf("DATABASE_USER environment variable is required")
	}
	if config.Database.Password == "" {
		return nil, fmt.Errorf("DATABASE_PASSWORD environment variable is required")
	}
	if config.Database.Name == "" {
		return nil, fmt.Errorf("DATABASE_NAME environment variable is required")
	}
	if config.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	log.Println("Configuration loaded successfully from environment variables")
	return config, nil
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}