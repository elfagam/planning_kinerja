package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config contains runtime configuration loaded from environment variables.
type Config struct {
	AppName                  string
	AppEnv                   string
	HTTPAddr                 string
	MySQLDSN                 string
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPass                   string
	DBName                   string
	DBConnectMaxRetries      int
	DBConnectRetryDelaySecs  int
	DBMaxOpenConns           int
	DBMaxIdleConns           int
	DBConnMaxLifetimeMins    int
	DBConnMaxIdleTimeMins    int
	GinMode                  string
	LogLevel                 string
	AuthEnabled              bool
	AuthToken                string
	DevAuthUserEmail         string
	JWTIssuer                string
	JWTAccessTokenTTLMinutes int
	JWTRefreshTokenTTLHours  int
	ReadTimeoutSeconds       int
	WriteTimeoutSeconds      int
	ShutdownTimeoutSeconds   int
}

func Load() *Config {
	loadDotEnvIfExists(".env")

	dbHost := getenv("DB_HOST", "localhost")
	dbPort := getenv("DB_PORT", "3306")
	dbUser := getenv("DB_USER", "root")
	dbPass := getenv("DB_PASSWORD", "")
	dbName := getenv("DB_NAME", "e-plan-ai")

	// Build fallback DSN from individual components
	defaultDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	return &Config{
		AppName:                  getenv("APP_NAME", "e-plan-ai"),
		AppEnv:                   getenv("APP_ENV", "development"),
		HTTPAddr:                 getenv("HTTP_ADDR", ":8080"),
		MySQLDSN:                 getenv("MYSQL_DSN", defaultDSN),
		DBHost:                   dbHost,
		DBPort:                   dbPort,
		DBUser:                   dbUser,
		DBPass:                   dbPass,
		DBName:                   dbName,
		DBConnectMaxRetries:      getenvInt("DB_CONNECT_MAX_RETRIES", 10),
		DBConnectRetryDelaySecs:  getenvInt("DB_CONNECT_RETRY_DELAY_SECONDS", 2),
		DBMaxOpenConns:           getenvInt("DB_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns:           getenvInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetimeMins:    getenvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30),
		DBConnMaxIdleTimeMins:    getenvInt("DB_CONN_MAX_IDLE_TIME_MINUTES", 10),
		GinMode:                  getenv("GIN_MODE", "debug"),
		LogLevel:                 getenv("LOG_LEVEL", "info"),
		AuthEnabled:              getenvBool("AUTH_ENABLED", true),
		AuthToken:                getenv("AUTH_TOKEN", "change-me-in-production"),
		DevAuthUserEmail:         getenv("DEV_AUTH_USER_EMAIL", "superadmin@rsudcontoh.go.id"),
		JWTIssuer:                getenv("JWT_ISSUER", "e-plan-ai"),
		JWTAccessTokenTTLMinutes: getenvInt("JWT_ACCESS_TOKEN_TTL_MINUTES", 15),
		JWTRefreshTokenTTLHours:  getenvInt("JWT_REFRESH_TOKEN_TTL_HOURS", 24),
		ReadTimeoutSeconds:       getenvInt("READ_TIMEOUT_SECONDS", 10),
		WriteTimeoutSeconds:      getenvInt("WRITE_TIMEOUT_SECONDS", 10),
		ShutdownTimeoutSeconds:   getenvInt("SHUTDOWN_TIMEOUT_SECONDS", 10),
	}
}

func loadDotEnvIfExists(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
}

func getenvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}

	return b
}
