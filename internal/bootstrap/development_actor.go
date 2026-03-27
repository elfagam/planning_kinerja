package bootstrap

import (
	"log"
	"strings"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/middleware"

	"gorm.io/gorm"
)

type developmentActorRow struct {
	ID       uint64
	Email    string
	FullName string
	Role     string
}

func resolveDevelopmentActor(cfg *config.Config) middleware.DevelopmentActorContext {
	if cfg.AuthEnabled || !strings.EqualFold(strings.TrimSpace(cfg.AppEnv), "development") {
		return middleware.DevelopmentActorContext{}
	}

	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("development actor unavailable: %v", err)
		return middleware.DevelopmentActorContext{}
	}

	preferredEmail := strings.ToLower(strings.TrimSpace(cfg.DevAuthUserEmail))
	actor, err := lookupDevelopmentActor(db, preferredEmail)
	if err != nil {
		log.Printf("development actor lookup failed: %v", err)
		return middleware.DevelopmentActorContext{}
	}
	if actor.ID == 0 && preferredEmail != "" {
		actor, err = lookupDevelopmentActor(db, "")
		if err != nil {
			log.Printf("development actor fallback lookup failed: %v", err)
			return middleware.DevelopmentActorContext{}
		}
	}
	if actor.ID == 0 {
		return middleware.DevelopmentActorContext{}
	}

	return middleware.DevelopmentActorContext{
		UserID:   actor.ID,
		Email:    actor.Email,
		FullName: actor.FullName,
		Role:     actor.Role,
	}
}

func lookupDevelopmentActor(db *gorm.DB, preferredEmail string) (developmentActorRow, error) {
	queries := []string{
		`SELECT
			u.id,
			u.email,
			COALESCE(NULLIF(u.nama_lengkap, ''), u.email) AS full_name,
			COALESCE(NULLIF(u.role, ''), 'PERENCANA') AS role
		FROM users u
		WHERE u.aktif = 1
		  AND (? = '' OR LOWER(u.email) = ?)
		ORDER BY CASE COALESCE(NULLIF(u.role, ''), 'PERENCANA')
		  WHEN 'OPERATOR' THEN 1
		  WHEN 'PERENCANA' THEN 2
		  WHEN 'ADMIN' THEN 3
		  WHEN 'VERIFIKATOR' THEN 4
		  WHEN 'PIMPINAN' THEN 5
		  ELSE 6
		END,
		u.id ASC
		LIMIT 1`,
		`SELECT
			u.id,
			u.email,
			COALESCE(NULLIF(u.full_name, ''), u.email) AS full_name,
			COALESCE(NULLIF(u.role, ''), 'PERENCANA') AS role
		FROM users u
		WHERE u.is_active = 1
		  AND (? = '' OR LOWER(u.email) = ?)
		ORDER BY CASE COALESCE(NULLIF(u.role, ''), 'PERENCANA')
		  WHEN 'OPERATOR' THEN 1
		  WHEN 'PERENCANA' THEN 2
		  WHEN 'ADMIN' THEN 3
		  WHEN 'VERIFIKATOR' THEN 4
		  WHEN 'PIMPINAN' THEN 5
		  ELSE 6
		END,
		u.id ASC
		LIMIT 1`,
		`SELECT
			u.id,
			u.email,
			COALESCE(NULLIF(u.full_name, ''), u.email) AS full_name,
			'PERENCANA' AS role
		FROM users u
		WHERE u.is_active = 1
		  AND (? = '' OR LOWER(u.email) = ?)
		ORDER BY u.id ASC
		LIMIT 1`,
	}

	var lastErr error
	for _, query := range queries {
		var row developmentActorRow
		tx := db.Raw(query, preferredEmail, preferredEmail).Scan(&row)
		if tx.Error != nil {
			lastErr = tx.Error
			continue
		}
		if tx.RowsAffected > 0 {
			return row, nil
		}
	}

	if lastErr != nil {
		return developmentActorRow{}, lastErr
	}

	return developmentActorRow{}, nil
}
