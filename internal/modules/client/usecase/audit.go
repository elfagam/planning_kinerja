package usecase

import (
	"encoding/json"
	"strings"
)

const clientResourceType = "CLIENT"

func buildAuditEntry(actor Actor, action string, resourceID uint64, payload map[string]any) AuditLogEntry {
	var userIDPtr *uint64
	if actor.ID > 0 {
		uid := actor.ID
		userIDPtr = &uid
	}

	resourceIDPtr := resourceIDPointer(resourceID)
	payloadPtr := jsonPayloadPointer(payload)

	return AuditLogEntry{
		UserID:         userIDPtr,
		Action:         action,
		ResourceType:   clientResourceType,
		ResourceID:     resourceIDPtr,
		RequestPayload: payloadPtr,
		IPAddress:      nonEmptyStringPointer(actor.IPAddress),
		UserAgent:      nonEmptyStringPointer(actor.UserAgent),
	}
}

func resourceIDPointer(id uint64) *uint64 {
	if id == 0 {
		return nil
	}
	v := id
	return &v
}

func jsonPayloadPointer(payload map[string]any) *string {
	if len(payload) == 0 {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

func nonEmptyStringPointer(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
