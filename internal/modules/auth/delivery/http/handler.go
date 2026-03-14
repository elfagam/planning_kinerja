package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/middleware"
	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	cfg        config.Config
	db         *gorm.DB
	signingKey []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type loginRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=150"`
	Email    string `json:"email" binding:"omitempty,email,max=150"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=20"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type systemUser struct {
	ID           uint64
	Email        string
	FullName     string
	PasswordHash string
	IsActive     bool
	Role         string
}

func NewHandler(cfg config.Config) *Handler {
	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		db = nil
	}

	key := cfg.AuthToken
	if key == "" {
		key = "change-me-in-production"
	}

	accessTTL := time.Duration(cfg.JWTAccessTokenTTLMinutes) * time.Minute
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}

	refreshTTL := time.Duration(cfg.JWTRefreshTokenTTLHours) * time.Hour
	if refreshTTL <= 0 {
		refreshTTL = 24 * time.Hour
	}

	return &Handler{
		cfg:        cfg,
		db:         db,
		signingKey: []byte(key),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.login)
	rg.POST("/refresh", h.refresh)
	rg.POST("/logout", h.logout)

	protected := rg.Group("")
	protected.Use(middleware.Auth(true, string(h.signingKey)))
	protected.GET("/me", h.me)
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, bindErrorMessage(err, "invalid request payload"))
		return
	}

	identifier := strings.TrimSpace(req.Email)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Username)
	}

	if identifier == "" || strings.TrimSpace(req.Password) == "" {
		response.Error(c, http.StatusBadRequest, "username/email and password are required")
		return
	}

	user, err := h.findSystemUserByEmail(identifier)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "database connection is unavailable") {
			response.Error(c, http.StatusInternalServerError, "failed to authenticate user")
			return
		}

		// Do not leak schema-specific lookup errors to clients.
		response.Error(c, http.StatusUnauthorized, "invalid username or password")
		return
	}
	if user == nil {
		response.Error(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if !user.IsActive {
		response.Error(c, http.StatusForbidden, "user is not active")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	accessToken, expiresIn, err := h.generateToken(user, "access", h.accessTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	refreshToken, _, err := h.generateToken(user, "refresh", h.refreshTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	response.Success(c, tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	})
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, bindErrorMessage(err, "invalid request payload"))
		return
	}

	claims, err := h.parseToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	if claims.TokenType != "refresh" {
		response.Error(c, http.StatusUnauthorized, "invalid token type")
		return
	}

	user := &systemUser{
		ID:       claims.UserID,
		Email:    claims.Email,
		FullName: claims.FullName,
		Role:     claims.Role,
	}
	if user.Email == "" {
		user.Email = claims.Username
	}

	accessToken, expiresIn, err := h.generateToken(user, "access", h.accessTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	refreshToken, _, err := h.generateToken(user, "refresh", h.refreshTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	response.Success(c, tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	})
}

func (h *Handler) logout(c *gin.Context) {
	// Stateless JWT logout is handled on client by deleting stored tokens.
	response.Success(c, gin.H{"message": "logged out"})
}

func (h *Handler) me(c *gin.Context) {
	userID, _ := c.Get("auth.user_id")
	username, _ := c.Get("auth.username")
	email, _ := c.Get("auth.email")
	fullName, _ := c.Get("auth.full_name")
	role, _ := c.Get("auth.role")
	subject, _ := c.Get("auth.subject")

	response.Success(c, gin.H{
		"user_id":   userID,
		"username":  username,
		"email":     email,
		"full_name": fullName,
		"role":      role,
		"subject":   subject,
	})
}

func (h *Handler) generateToken(user *systemUser, tokenType string, ttl time.Duration) (string, int64, error) {
	now := time.Now()
	exp := now.Add(ttl)
	username := user.Email
	if username == "" {
		username = user.FullName
	}

	claims := middleware.Claims{
		UserID:    user.ID,
		Username:  username,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			Issuer:    h.cfg.JWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.signingKey)
	if err != nil {
		return "", 0, err
	}

	return tokenString, int64(ttl.Seconds()), nil
}

func (h *Handler) parseToken(tokenString string) (*middleware.Claims, error) {
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (h *Handler) findSystemUserByEmail(identifier string) (*systemUser, error) {
	if h.db == nil {
		return nil, fmt.Errorf("database connection is unavailable")
	}

	email := strings.ToLower(strings.TrimSpace(identifier))
	if email == "" {
		return nil, nil
	}

	queries := []string{
		// Legacy schema from SQL migrations (users.full_name, users.is_active + user_roles table)
		`
		SELECT
			u.id,
			u.email,
			u.full_name AS full_name,
			u.password_hash,
			u.is_active AS is_active,
			COALESCE(r.code, 'PERENCANA') AS role
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE LOWER(u.email) = ?
		LIMIT 1
		`,
		// Legacy-compatible schema without user_roles table.
		`
		SELECT
			u.id,
			u.email,
			u.full_name AS full_name,
			u.password_hash,
			u.is_active AS is_active,
			'PERENCANA' AS role
		FROM users u
		WHERE LOWER(u.email) = ?
		LIMIT 1
		`,
		// Newer schema from gorm models (users.nama_lengkap, users.aktif, users.role)
		`
		SELECT
			u.id,
			u.email,
			u.nama_lengkap AS full_name,
			u.password_hash,
			u.aktif AS is_active,
			u.role AS role
		FROM users u
		WHERE LOWER(u.email) = ?
		LIMIT 1
		`,
		// Hybrid schema: legacy name/active with inline role column.
		`
		SELECT
			u.id,
			u.email,
			u.full_name AS full_name,
			u.password_hash,
			u.is_active AS is_active,
			u.role AS role
		FROM users u
		WHERE LOWER(u.email) = ?
		LIMIT 1
		`,
	}

	var lastErr error
	for _, query := range queries {
		candidate := systemUser{}
		tx := h.db.Raw(query, email).Scan(&candidate)
		if tx.Error != nil {
			lastErr = tx.Error
			continue
		}
		if tx.RowsAffected > 0 {
			if strings.TrimSpace(candidate.Role) == "" {
				candidate.Role = "PERENCANA"
			}
			return &candidate, nil
		}
	}

	if lastErr != nil {
		msg := strings.ToLower(lastErr.Error())
		// Common schema mismatch errors should not break authentication flow.
		if strings.Contains(msg, "unknown column") ||
			strings.Contains(msg, "doesn't exist") ||
			strings.Contains(msg, "no such table") {
			return nil, nil
		}
		return nil, lastErr
	}

	return nil, nil
}

func bindErrorMessage(err error, fallback string) string {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) || len(validationErrs) == 0 {
		return fallback
	}

	fe := validationErrs[0]
	field := validationFieldName(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s minimum length is %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s maximum length is %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	default:
		return fallback
	}
}

func validationFieldName(raw string) string {
	switch raw {
	case "Username":
		return "username"
	case "Email":
		return "email"
	case "Password":
		return "password"
	case "RefreshToken":
		return "refresh_token"
	default:
		return strings.ToLower(raw)
	}
}
