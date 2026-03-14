package http

type upsertClientRequest struct {
	Kode           string  `json:"kode" binding:"required,min=2,max=50"`
	Nama           string  `json:"nama" binding:"required,min=3,max=255"`
	UnitPengusulID *uint64 `json:"unit_pengusul_id" binding:"omitempty,gt=0"`
}

type transitionRequest struct {
	Reason string `json:"reason" binding:"omitempty,max=1000"`
	Note   string `json:"note" binding:"omitempty,max=1000"`
}
