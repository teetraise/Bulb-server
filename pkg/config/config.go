package config

import (
	"fmt"
	"log"

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
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	// Установка значений по умолчанию
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("jwt.expiresin", 24) // 24 часа

	// Чтение файла конфигурации
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Преобразование конфигурации в структуру
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	log.Println("Configuration loaded successfully")
	return config, nil
}
