package domain

import (
	"errors"
	"time"
)

// ExportIndikatorCSVFlatDTO is a flat DTO for exporting Indikator Kinerja CSV.
type ExportIndikatorCSVFlatDTO struct {
	// Top Layout
	ProgramNama              string
	KegiatanNama             string
	SubKegiatanNama          string
	UnitPengusulKode         string
	UnitPengusulNama         string
	RencanaKerjaTahun        int16

	// Detail Data
	RencanaKerjaKode         string
	RencanaKerjaNama         string
	StandarHargaIdRekening   string
	StandarHargaId           uint64
	IndikatorKode            string
	IndikatorNama            string
	HargaSatuan              float64
	Satuan                   string
	TargetTahunan            float64
	AnggaranTahunan          float64

	// Bottom Layout
	JabatanPenanggungJawab   string
	NamaPenanggungJawab      string
	NipPenanggungJawab       string
}

var (
	ErrInvalidTransition      = errors.New("invalid renja status transition")
	ErrRejectionReasonMissing = errors.New("rejection reason is required")
)

type Status string

const (
	StatusDraft     Status = "DRAFT"
	StatusSubmitted Status = "DIAJUKAN"
	StatusApproved  Status = "DISETUJUI"
	StatusRejected  Status = "DITOLAK"
)

// RencanaKerja represents domain state for Renja workflow.
type RencanaKerja struct {
	ID                     int64
	IndikatorSubKegiatanID int64
	Code                   string
	Name                   string
	Tahun                  int16
	Triwulan               int8
	UnitPengusulID         int64
	Status                 Status
	Notes                  string
	RejectedReason         string
	CreatedBy              int64
	UpdatedBy              int64
	ApprovedBy             int64
	ApprovedAt             time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// Renja is kept as an alias for backward compatibility in use cases.
type Renja = RencanaKerja

type StandarHarga struct {
	ID           uint64
	JenisStandar *string
	UraianBarang *string
	Spesifikasi  *string
	Satuan       *string
	HargaSatuan  *float64
	IdRekening   *string
}

// IndikatorRencanaKerja represents KPI and budgeting target for one rencana kerja.
type IndikatorRencanaKerja struct {
	ID               int64
	RencanaKerjaID   int64
	TbStandarHargaID *uint64
	StandarHarga     *StandarHarga
	Code             string
	Name            string
	Unit            string
	TargetTahunan   float64
	HargaSatuan     float64
	AnggaranTahunan float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RealisasiRencanaKerja represents periodical realization of one indikator rencana kerja.
type RealisasiRencanaKerja struct {
	ID                      int64
	IndikatorRencanaKerjaID int64
	Tahun                   int16
	Bulan                   int8
	Triwulan                int8
	NilaiRealisasi          float64
	RealisasiAnggaran       float64
	Keterangan              string
	DiinputOleh             int64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// Ajukan moves draft plan into submitted state.
func (r *RencanaKerja) Ajukan(actorID int64) error {
	if r.Status != StatusDraft {
		return ErrInvalidTransition
	}
	r.Status = StatusSubmitted
	r.RejectedReason = ""
	r.ApprovedBy = 0
	r.UpdatedBy = actorID
	return nil
}

// Setujui approves a submitted plan.
func (r *RencanaKerja) Setujui(actorID int64) error {
	if r.Status != StatusSubmitted {
		return ErrInvalidTransition
	}
	r.Status = StatusApproved
	r.RejectedReason = ""
	r.ApprovedBy = actorID
	r.UpdatedBy = actorID
	return nil
}

// Tolak rejects a submitted plan and requires rejection notes.
func (r *RencanaKerja) Tolak(actorID int64, reason string) error {
	if r.Status != StatusSubmitted {
		return ErrInvalidTransition
	}
	if reason == "" {
		return ErrRejectionReasonMissing
	}
	r.Status = StatusRejected
	r.RejectedReason = reason
	r.Notes = reason
	r.ApprovedBy = 0
	r.UpdatedBy = actorID
	return nil
}

// Submit keeps old API compatibility.
func (r *RencanaKerja) Submit(actorID int64) error { return r.Ajukan(actorID) }

// Approve keeps old API compatibility.
func (r *RencanaKerja) Approve(actorID int64) error { return r.Setujui(actorID) }

// Reject keeps old API compatibility.
func (r *RencanaKerja) Reject(actorID int64, reason string) error { return r.Tolak(actorID, reason) }
