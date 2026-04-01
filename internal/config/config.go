package config

import (
	"bufio"
	"fmt"
	"net/url"
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
	// 1. Muat file .env jika ada (berguna saat jalan di lokal/laptop)
	loadDotEnvIfExists(".env")

	// 2. Ambil Variabel Database dengan sistem Prioritas (Railway -> Lokal -> Default)
	// Railway standar: MYSQLHOST, MYSQLPORT, MYSQLUSER, MYSQLPASSWORD, MYSQLDATABASE
	// Beberapa sistem lain: DB_HOST, DB_USER, DB_PASSWORD, dll.
	dbHost := getenv("DB_HOST", getenv("MYSQLHOST", "localhost"))
	dbPort := getenv("DB_PORT", getenv("MYSQLPORT", "3306"))
	dbUser := getenv("DB_USER", getenv("MYSQLUSER", "root"))
	dbPass := getenv("DB_PASSWORD", getenv("MYSQLPASSWORD", ""))
	dbName := getenv("DB_NAME", getenv("MYSQLDATABASE", "e-plan-ai"))

	// 3. Rakit DSN (Data Source Name)
	dsn := os.Getenv("MYSQL_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}

	// Jika DSN dalam format URI (mysql://), konversi ke format DSN driver Go-MySQL
	if strings.HasPrefix(dsn, "mysql://") {
		parsedDSN, err := convertURIToDSN(dsn)
		if err == nil {
			dsn = parsedDSN
		}
	}

	if dsn == "" {
		// Jika DSN masih kosong, rakit dari komponen individual
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPass, dbHost, dbPort, dbName)
	} else {
		// Pastikan parseTime=True ada untuk GORM
		if !strings.Contains(dsn, "parseTime=") {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn += sep + "parseTime=True"
		}
	}

	// 4. Diagnostic Logging (Membantu kita melihat apa yang terbaca di Railway Logs)
	if os.Getenv("RAILWAY_ENVIRONMENT_NAME") != "" || os.Getenv("PORT") != "" {
		fmt.Printf("[CONFIG] Railway environment detected\n")
		fmt.Printf("[CONFIG] Target DB Host: %s:%s\n", dbHost, dbPort)
		fmt.Printf("[CONFIG] Target DB User: %s\n", dbUser)
		if dsn != "" {
			// Sembunyikan password jika ada di DSN saat logging
			maskedDSN := dsn
			if atIdx := strings.Index(dsn, "@"); atIdx > 0 {
				if colonIdx := strings.Index(dsn[:atIdx], ":"); colonIdx > 0 {
					maskedDSN = dsn[:colonIdx+1] + "****" + dsn[atIdx:]
				}
			}
			fmt.Printf("[CONFIG] Using connection string (DSN): %s\n", maskedDSN)
		}
	}

	return &Config{
		AppName:                  getenv("APP_NAME", "e-plan-ai"),
		AppEnv:                   getenv("APP_ENV", "development"),
		HTTPAddr:                 getenv("HTTP_ADDR", ":8080"),
		MySQLDSN:                 dsn,
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
		AuthToken:                getenv("AUTH_TOKEN", "rahasia-super-kuat-123"), // Default fallback agar JWT tidak crash
		DevAuthUserEmail:         getenv("DEV_AUTH_USER_EMAIL", "superadmin@rsudcontoh.go.id"),
		JWTIssuer:                getenv("JWT_ISSUER", "e-plan-ai"),
		JWTAccessTokenTTLMinutes: getenvInt("JWT_ACCESS_TOKEN_TTL_MINUTES", 15),
		JWTRefreshTokenTTLHours:  getenvInt("JWT_REFRESH_TOKEN_TTL_HOURS", 24),
		ReadTimeoutSeconds:       getenvInt("READ_TIMEOUT_SECONDS", 10),
		WriteTimeoutSeconds:      getenvInt("WRITE_TIMEOUT_SECONDS", 10),
		ShutdownTimeoutSeconds:   getenvInt("SHUTDOWN_TIMEOUT_SECONDS", 10),
	}
}

// --- FUNGSI HELPER ---

func loadDotEnvIfExists(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // File tidak ada, abaikan saja (misal saat di production)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Abaikan baris kosong atau komentar
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		
		// Jangan timpa variabel yang sudah ada di environment OS
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
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

// convertURIToDSN mengubah format URI (mysql://user:pass@host:port/db) 
// ke format DSN yang dimengerti driver go-sql-driver/mysql (user:pass@tcp(host:port)/db)
func convertURIToDSN(uriStr string) (string, error) {
	u, err := url.Parse(uriStr)
	if err != nil {
		return "", err
	}

	if u.Scheme != "mysql" {
		return uriStr, nil
	}

	user := u.User.Username()
	pass, _ := u.User.Password()
	host := u.Host
	db := strings.TrimPrefix(u.Path, "/")
	
	// Default MySQL port jika tidak ada di host
	if !strings.Contains(host, ":") {
		host = host + ":3306"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, db)
	
	// Tambahkan query parameters jika ada
	if u.RawQuery != "" {
		dsn += "?" + u.RawQuery
	}

	return dsn, nil
}
