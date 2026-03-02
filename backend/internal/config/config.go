package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	MinIO    MinIOConfig
	JWT      JWTConfig
	Auth     AuthConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Port         string
	Mode         string // debug | release
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	DefaultBucket   string
	PublicEndpoint  string // CDN or public MinIO URL
	SignedURLExpiry time.Duration
}


type JWTConfig struct {
	Secret   string
	Expiry   time.Duration
	Required bool // false = no auth enforced (open API)
}

type AuthConfig struct {
	DefaultAdminUsername string
	DefaultAdminPassword string
}

type UploadConfig struct {
	MaxFileSizeBytes int64
	AllowedMIMEs     []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  parseDuration(getEnv("SERVER_READ_TIMEOUT", "30s")),
			WriteTimeout: parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "60s")),
		},
		Database: DatabaseConfig{
			DSN:          buildDSN(),
			MaxOpenConns: parseInt(getEnv("DB_MAX_OPEN_CONNS", "25")),
			MaxIdleConns: parseInt(getEnv("DB_MAX_IDLE_CONNS", "5")),
			MaxLifetime:  parseDuration(getEnv("DB_MAX_LIFETIME", "5m")),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnv("MINIO_USE_SSL", "false") == "true",
			DefaultBucket:   getEnv("MINIO_DEFAULT_BUCKET", "media"),
			PublicEndpoint:  getEnv("MINIO_PUBLIC_ENDPOINT", "http://localhost:9000"),
			SignedURLExpiry: parseDuration(getEnv("MINIO_SIGNED_URL_EXPIRY", "1h")),
		},

		JWT: JWTConfig{
			Secret:   getEnv("JWT_SECRET", "change-me-to-a-long-random-secret-at-least-32-chars"),
			Expiry:   parseDuration(getEnv("JWT_EXPIRY", "24h")),
			Required: getEnv("AUTH_REQUIRED", "false") == "true",
		},
		Auth: AuthConfig{
			DefaultAdminUsername: getEnv("AUTH_ADMIN_USERNAME", "admin"),
			DefaultAdminPassword: getEnv("AUTH_ADMIN_PASSWORD", "admin123456"),
		},
		Upload: UploadConfig{
			MaxFileSizeBytes: parseInt64(getEnv("UPLOAD_MAX_SIZE_BYTES", "104857600")), // 100MB
			AllowedMIMEs:     nil,                                                      // nil = allow all; set to restrict
		},
	}

	return cfg, nil
}

func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "postgres")
	name := getEnv("DB_NAME", "media_cms")
	sslmode := getEnv("DB_SSLMODE", "disable")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, sslmode)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %q is not set", key))
	}
	return v
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}
