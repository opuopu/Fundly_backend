package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	Environment string
	AutoMigrate bool

	UploadsDir   string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string
}

// SMTPConfigured reports whether enough SMTP settings are present to send real email.
func (c *Config) SMTPConfigured() bool {
	return c.SMTPHost != "" && c.SMTPUsername != "" && c.SMTPPassword != ""
}

func Load() *Config {
	loadDotEnv(".env")

	port := getEnvOrDefault("PORT", "8081")
	env := getEnvOrDefault("APP_ENV", "development")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := getEnvOrDefault("DB_HOST", "localhost")
		dbPort := getEnvOrDefault("DB_PORT", "5435")
		user := getEnvOrDefault("DB_USER", "fundly")
		pass := getEnvOrDefault("DB_PASSWORD", "fundly123")
		dbName := getEnvOrDefault("DB_NAME", "fundly")
		sslMode := getEnvOrDefault("DB_SSLMODE", "disable")

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, dbPort, dbName, sslMode)
	}

	autoMigrate := os.Getenv("AUTO_MIGRATE") != "false" // default to true

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		Environment: env,
		AutoMigrate: autoMigrate,
		UploadsDir:  getEnvOrDefault("UPLOADS_DIR", "./uploads"),

		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		SMTPFrom:     os.Getenv("SMTP_FROM"),
		SMTPFromName: getEnvOrDefault("SMTP_FROM_NAME", "Fundly"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if !strings.HasPrefix(v, `"`) && !strings.HasPrefix(v, `'`) {
				if idx := strings.Index(v, " #"); idx != -1 {
					v = strings.TrimSpace(v[:idx])
				} else if idx := strings.Index(v, "\t#"); idx != -1 {
					v = strings.TrimSpace(v[:idx])
				}
			}
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}
