package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
)

var ErrDatabaseConnection = errors.New("database connection failure")

// ConnectionError marks failures related to establishing/using DB connections.
type ConnectionError struct {
	Operation string
	Err       error
}

func (e *ConnectionError) Error() string {
	if e == nil {
		return ErrDatabaseConnection.Error()
	}
	if e.Operation == "" {
		return fmt.Sprintf("%s: %v", ErrDatabaseConnection.Error(), e.Err)
	}
	return fmt.Sprintf("%s (%s): %v", ErrDatabaseConnection.Error(), e.Operation, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *ConnectionError) Is(target error) bool {
	return target == ErrDatabaseConnection
}

func IsConnectionError(err error) bool {
	return errors.Is(err, ErrDatabaseConnection)
}

// WrapConnectionError wraps an error as a database connection failure.
func WrapConnectionError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if IsConnectionError(err) {
		return err
	}
	return &ConnectionError{Operation: operation, Err: err}
}

// WrapIfConnectionError wraps only known connection-related errors.
func WrapIfConnectionError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if IsConnectionError(err) {
		return err
	}
	if !looksLikeConnectionError(err) {
		return err
	}
	return &ConnectionError{Operation: operation, Err: err}
}

func looksLikeConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrConnDone) || errors.Is(err, driver.ErrBadConn) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr)
}

func ConnectionFailureHint() string {
	return "check MYSQL_DSN, MySQL service status, network/port, and DB credentials"
}
