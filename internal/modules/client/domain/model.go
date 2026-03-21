package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidStatus      = errors.New("invalid client status")
	ErrInvalidTransition  = errors.New("invalid client status transition")
	ErrValidation         = errors.New("validation error")
	ErrReasonRequired     = errors.New("reason is required")
	ErrClientNotFound     = errors.New("client not found")
	ErrForbiddenOperation = errors.New("forbidden client operation")
)

type Status string

const (
	StatusDraft     Status = "DRAFT"
	StatusDiajukan  Status = "DIAJUKAN"
	StatusDisetujui Status = "DISETUJUI"
	StatusDitolak   Status = "DITOLAK"
)

type Client struct {
	ID             uint64
	Kode           string
	Nama           string
	Status         Status
	UnitPengusulID *uint64
	CreatedBy      *uint64
	UpdatedBy      *uint64
	ApprovedBy     *uint64
	ApprovedAt     *time.Time
	RejectedBy     *uint64
	RejectedAt     *time.Time
	RejectedReason *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type StatusHistory struct {
	ID         uint64
	ClientID   uint64
	FromStatus *Status
	ToStatus   Status
	Action     string
	Reason     *string
	Note       *string
	ActorID    *uint64
	ActorName  *string
	CreatedAt  time.Time
}

func ParseStatus(raw string) (Status, error) {
	s := Status(strings.ToUpper(strings.TrimSpace(raw)))
	switch s {
	case StatusDraft, StatusDiajukan, StatusDisetujui, StatusDitolak:
		return s, nil
	default:
		return "", ErrInvalidStatus
	}
}

func CanTransition(from, to Status) bool {
	switch from {
	case StatusDraft:
		return to == StatusDiajukan
	case StatusDiajukan:
		return to == StatusDraft || to == StatusDitolak || to == StatusDisetujui
	case StatusDitolak:
		return to == StatusDraft
	default:
		return false
	}
}
