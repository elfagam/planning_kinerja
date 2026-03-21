package domain

func ActionTargetStatus(action string) (Status, bool) {
	switch action {
	case "submit":
		return StatusDiajukan, true
	case "unsubmit":
		return StatusDraft, true
	case "reject":
		return StatusDitolak, true
	case "re-evaluate":
		return StatusDraft, true
	case "approve":
		return StatusDisetujui, true
	default:
		return "", false
	}
}
