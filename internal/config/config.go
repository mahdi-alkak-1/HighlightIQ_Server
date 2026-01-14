package config

import "os"

// MySQLConfig holds DB connection settings.
type MySQLConfig struct {
	Host string
	Port string
	Name string
	User string
	Pass string
}

type Config struct {
	MySQL                 MySQLConfig
	JWTSecret             string
	RecordingsDir         string
	N8NWebhookSecret      string
	ClipsBaseURL          string
	N8NPublishWebhookURL  string
	N8NPublishWebhookAuth string
}

// Load reads configuration from environment variables with sane defaults.
func Load() Config {
	return Config{
		MySQL: MySQLConfig{
			Host: getenv("DB_HOST", "127.0.0.1"),
			Port: getenv("DB_PORT", "3307"),
			Name: getenv("DB_NAME", "highlightiq"),
			User: getenv("DB_USER", "highlightiq"),
			Pass: getenv("DB_PASS", "highlightiq_pass"),
		},
		JWTSecret:             getenv("JWT_SECRET", "dev-secret-change-me"),
		RecordingsDir:         getenv("RECORDINGS_DIR", "D:\\recordings"),
		N8NWebhookSecret:      getenv("N8N_WEBHOOK_SECRET", ""),
		ClipsBaseURL:          getenv("CLIPS_BASE_URL", ""),
		N8NPublishWebhookURL:  getenv("N8N_PUBLISH_WEBHOOK_URL", ""),
		N8NPublishWebhookAuth: getenv("N8N_PUBLISH_WEBHOOK_AUTH", ""),
	}
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
