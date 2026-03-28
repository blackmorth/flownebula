package api

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerListenAddr       string
	ServerMetricsAddr      string
	DBPath                 string
	CORSAllowOrigins       string
	CORSAllowMethods       string
	CORSAllowHeaders       string
	CORSAllowCredentials   bool
	UploadRatePerMinute    int
	UploadMaxPayloadBytes  int
	SessionRetentionInDays int
}

func LoadConfig() Config {
	return Config{
		ServerListenAddr:       getEnv("SERVER_LISTEN_ADDR", ":8080"),
		ServerMetricsAddr:      getEnv("SERVER_METRICS_ADDR", ":9109"),
		DBPath:                 getEnv("DB_PATH", "nebula.db"),
		CORSAllowOrigins:       getEnv("CORS_ALLOW_ORIGINS", "http://localhost:8081"),
		CORSAllowMethods:       getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
		CORSAllowHeaders:       getEnv("CORS_ALLOW_HEADERS", "Content-Type, Authorization"),
		CORSAllowCredentials:   getEnvBool("CORS_ALLOW_CREDENTIALS", true),
		UploadRatePerMinute:    getEnvInt("UPLOAD_RATE_PER_MINUTE", 60),
		UploadMaxPayloadBytes:  getEnvInt("UPLOAD_MAX_PAYLOAD_BYTES", 2*1024*1024),
		SessionRetentionInDays: getEnvInt("SESSION_RETENTION_DAYS", 0),
	}
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}
