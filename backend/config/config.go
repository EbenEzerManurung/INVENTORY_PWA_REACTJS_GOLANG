package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port    string
	GinMode string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT
	JWTSecret         string
	JWTRefreshSecret  string
	JWTExpiration     string
	JWTRefreshExpiration string

	// Seeder
	RunSeeder bool

	// Upload
	UploadDir   string
	MaxFileSize int64
}

// AppConfig is the global configuration instance
var AppConfig Config

// InitConfig initializes configuration from environment variables
func InitConfig() error {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found, using environment variables")
	}

	// Parse RunSeeder
	runSeeder := false
	if val := os.Getenv("RUN_SEEDER"); val == "true" || val == "1" {
		runSeeder = true
	}

	// Parse MaxFileSize
	maxFileSize := int64(2097152) // default 2MB
	if val := os.Getenv("MAX_FILE_SIZE"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			maxFileSize = parsed
		}
	}

	AppConfig = Config{
		// Server
		Port:    getEnv("PORT", "5000"),
		GinMode: getEnv("GIN_MODE", "debug"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "inventory_db"),

		// JWT
		JWTSecret:         getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		JWTRefreshSecret:  getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-in-production"),
		JWTExpiration:     getEnv("JWT_EXPIRATION", "24h"),
		JWTRefreshExpiration: getEnv("JWT_REFRESH_EXPIRATION", "168h"),

		// Seeder
		RunSeeder: runSeeder,

		// Upload
		UploadDir:   getEnv("UPLOAD_DIR", "uploads"),
		MaxFileSize: maxFileSize,
	}

	log.Println("✅ Configuration loaded successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}