package api

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	CORSAllowOrigins       string
	UploadRatePerMinute    int
	UploadMaxPayloadBytes  int
	SessionRetentionInDays int
}

func LoadConfig() Config {
	return Config{
		CORSAllowOrigins:       getEnv("CORS_ALLOW_ORIGINS", "http://localhost:8081"),
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
