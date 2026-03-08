package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"e-plan-ai/internal/modules/crud/domain"
	"e-plan-ai/internal/shared/database"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, database.WrapConnectionError("sql open", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, database.WrapConnectionError("sql ping", err)
	}

	s := &MySQLStore{db: db}
	if err := s.ensureTable(); err != nil {
		_ = db.Close()
		return nil, database.WrapIfConnectionError("ensure crud_records table", err)
	}

	return s, nil
}

func (s *MySQLStore) ensureTable() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS crud_records (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    resource VARCHAR(64) NOT NULL,
    code VARCHAR(50) NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    attributes JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_crud_records_resource (resource),
    INDEX idx_crud_records_name (name),
    INDEX idx_crud_records_code (code)
);`
	_, err := s.db.Exec(ddl)
	return err
}

func (s *MySQLStore) List(resource string, filter domain.ListFilter) ([]domain.Record, int64, error) {
	const countQuery = `
SELECT COUNT(1)
FROM crud_records
WHERE resource = ? AND (? = '' OR name LIKE ? OR code LIKE ?)`

	like := "%" + filter.Query + "%"
	var total int64
	if err := s.db.QueryRow(countQuery, resource, filter.Query, like, like).Scan(&total); err != nil {
		return nil, 0, database.WrapIfConnectionError("count crud records", err)
	}

	const listQuery = `
SELECT id, code, name, description, attributes, created_at, updated_at
FROM crud_records
WHERE resource = ? AND (? = '' OR name LIKE ? OR code LIKE ?)
ORDER BY id DESC
LIMIT ? OFFSET ?`

	rows, err := s.db.Query(listQuery, resource, filter.Query, like, like, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, database.WrapIfConnectionError("list crud records", err)
	}
	defer rows.Close()

	items := []domain.Record{}
	for rows.Next() {
		var item domain.Record
		var attributesRaw []byte
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &attributesRaw, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, database.WrapIfConnectionError("scan crud record", err)
		}
		item.Attributes = map[string]any{}
		if len(attributesRaw) > 0 {
			if err := json.Unmarshal(attributesRaw, &item.Attributes); err != nil {
				return nil, 0, fmt.Errorf("invalid attributes JSON: %w", err)
			}
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, database.WrapIfConnectionError("iterate crud records", err)
	}

	return items, total, nil
}

func (s *MySQLStore) Create(resource string, payload domain.Payload) (domain.Record, error) {
	attributesRaw, err := json.Marshal(payload.Attributes)
	if err != nil {
		return domain.Record{}, err
	}

	const insertQuery = `
INSERT INTO crud_records (resource, code, name, description, attributes)
VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.Exec(insertQuery, resource, payload.Code, payload.Name, payload.Description, attributesRaw)
	if err != nil {
		return domain.Record{}, database.WrapIfConnectionError("insert crud record", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Record{}, database.WrapIfConnectionError("read inserted crud record id", err)
	}

	return s.Get(resource, id)
}

func (s *MySQLStore) Get(resource string, id int64) (domain.Record, error) {
	const query = `
SELECT id, code, name, description, attributes, created_at, updated_at
FROM crud_records
WHERE resource = ? AND id = ?`

	var item domain.Record
	var attributesRaw []byte
	err := s.db.QueryRow(query, resource, id).Scan(&item.ID, &item.Code, &item.Name, &item.Description, &attributesRaw, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Record{}, ErrNotFound
		}
		return domain.Record{}, database.WrapIfConnectionError("get crud record", err)
	}
	item.Attributes = map[string]any{}
	if len(attributesRaw) > 0 {
		if err := json.Unmarshal(attributesRaw, &item.Attributes); err != nil {
			return domain.Record{}, fmt.Errorf("invalid attributes JSON: %w", err)
		}
	}
	return item, nil
}

func (s *MySQLStore) Update(resource string, id int64, payload domain.Payload) (domain.Record, error) {
	attributesRaw, err := json.Marshal(payload.Attributes)
	if err != nil {
		return domain.Record{}, err
	}

	const query = `
UPDATE crud_records
SET code = ?, name = ?, description = ?, attributes = ?, updated_at = CURRENT_TIMESTAMP
WHERE resource = ? AND id = ?`
	result, err := s.db.Exec(query, payload.Code, payload.Name, payload.Description, attributesRaw, resource, id)
	if err != nil {
		return domain.Record{}, database.WrapIfConnectionError("update crud record", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.Record{}, database.WrapIfConnectionError("read updated crud rows affected", err)
	}
	if affected == 0 {
		return domain.Record{}, ErrNotFound
	}
	return s.Get(resource, id)
}

func (s *MySQLStore) Delete(resource string, id int64) error {
	const query = `DELETE FROM crud_records WHERE resource = ? AND id = ?`
	result, err := s.db.Exec(query, resource, id)
	if err != nil {
		return database.WrapIfConnectionError("delete crud record", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return database.WrapIfConnectionError("read deleted crud rows affected", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
