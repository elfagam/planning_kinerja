package database

import (
	"context"
	"errors"
	"testing"
)

type fakeNetError struct{}

func (fakeNetError) Error() string   { return "network down" }
func (fakeNetError) Timeout() bool   { return false }
func (fakeNetError) Temporary() bool { return false }

func TestWrapConnectionErrorMarksAsConnectionFailure(t *testing.T) {
	baseErr := errors.New("dial tcp: connection refused")
	err := WrapConnectionError("gorm ping", baseErr)

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !IsConnectionError(err) {
		t.Fatal("expected wrapped error to be a connection error")
	}
	if !errors.Is(err, ErrDatabaseConnection) {
		t.Fatal("expected errors.Is(err, ErrDatabaseConnection) to be true")
	}

	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Fatal("expected error to unwrap to *ConnectionError")
	}
	if connErr.Operation != "gorm ping" {
		t.Fatalf("unexpected operation: got %q", connErr.Operation)
	}
}

func TestWrapIfConnectionErrorWithDeadlineExceeded(t *testing.T) {
	err := WrapIfConnectionError("sql ping", context.DeadlineExceeded)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !IsConnectionError(err) {
		t.Fatal("expected deadline exceeded to be treated as connection error")
	}
}

func TestWrapIfConnectionErrorWithNetError(t *testing.T) {
	err := WrapIfConnectionError("sql query", fakeNetError{})
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !IsConnectionError(err) {
		t.Fatal("expected net error to be treated as connection error")
	}
}

func TestWrapIfConnectionErrorLeavesDomainErrorsUntouched(t *testing.T) {
	baseErr := errors.New("duplicate key")
	err := WrapIfConnectionError("insert crud", baseErr)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if IsConnectionError(err) {
		t.Fatal("did not expect non-connection error to be tagged as connection error")
	}
	if !errors.Is(err, baseErr) {
		t.Fatal("expected original error to be preserved")
	}
}