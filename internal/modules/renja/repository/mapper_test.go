package repository

import (
	"testing"
	"time"

	"e-plan-ai/internal/modules/renja/domain"
	"e-plan-ai/internal/shared/database"
)

func TestRencanaKerjaMappingRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	d := domain.RencanaKerja{
		ID:                     10,
		IndikatorSubKegiatanID: 22,
		Code:                   "RK-001",
		Name:                   "Rencana Kerja A",
		Tahun:                  2026,
		Triwulan:               1,
		UnitPengusulID:         5,
		Status:                 domain.StatusRejected,
		RejectedReason:         "target tidak realistis",
		CreatedBy:              7,
		UpdatedBy:              8,
		ApprovedBy:             9,
		ApprovedAt:             now,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	gormModel := ToGormRencanaKerja(d)
	if gormModel.Triwulan == nil || *gormModel.Triwulan != 1 {
		t.Fatal("expected triwulan pointer to be populated")
	}
	if gormModel.DisetujuiOleh == nil || *gormModel.DisetujuiOleh != 9 {
		t.Fatal("expected disetujui_oleh pointer to be populated")
	}
	if gormModel.TanggalPersetujuan == nil {
		t.Fatal("expected tanggal_persetujuan pointer to be populated")
	}

	got := ToDomainRencanaKerja(gormModel)
	if got.ID != d.ID || got.IndikatorSubKegiatanID != d.IndikatorSubKegiatanID {
		t.Fatalf("identifier mapping mismatch: got=%+v want=%+v", got, d)
	}
	if got.Code != d.Code || got.Name != d.Name {
		t.Fatalf("text mapping mismatch: got=%+v want=%+v", got, d)
	}
	if got.Status != d.Status || got.RejectedReason != d.RejectedReason {
		t.Fatalf("status/catatan mapping mismatch: got=%+v want=%+v", got, d)
	}
}

func TestRencanaKerjaMappingHandlesNullableFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	gormModel := database.RencanaKerja{
		ID:                     1,
		IndikatorSubKegiatanID: 2,
		Kode:                   "RK-002",
		Nama:                   "Rencana Kerja B",
		Tahun:                  2026,
		Triwulan:               nil,
		UnitPengusulID:         3,
		Status:                 string(domain.StatusSubmitted),
		Catatan:                "catatan umum",
		DibuatOleh:             10,
		DisetujuiOleh:          nil,
		TanggalPersetujuan:     nil,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	got := ToDomainRencanaKerja(gormModel)
	if got.Triwulan != 0 || got.ApprovedBy != 0 || !got.ApprovedAt.IsZero() {
		t.Fatalf("nullable mapping mismatch: got triwulan=%d approvedBy=%d approvedAt=%v", got.Triwulan, got.ApprovedBy, got.ApprovedAt)
	}
}

func TestIndikatorMappingRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	d := domain.IndikatorRencanaKerja{
		ID:              10,
		RencanaKerjaID:  20,
		Code:            "IRK-001",
		Name:            "Indikator A",
		Unit:            "Persen",
		TargetTahunan:   99.5,
		AnggaranTahunan: 1250000,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	gormModel := ToGormIndikatorRencanaKerja(d)
	got := ToDomainIndikatorRencanaKerja(gormModel)

	if got.ID != d.ID || got.RencanaKerjaID != d.RencanaKerjaID {
		t.Fatalf("identifier mapping mismatch: got=%+v want=%+v", got, d)
	}
	if got.Code != d.Code || got.Name != d.Name || got.Unit != d.Unit {
		t.Fatalf("text mapping mismatch: got=%+v want=%+v", got, d)
	}
	if got.TargetTahunan != d.TargetTahunan || got.AnggaranTahunan != d.AnggaranTahunan {
		t.Fatalf("numeric mapping mismatch: got=%+v want=%+v", got, d)
	}
}

func TestRealisasiMappingHandlesNullablePeriod(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	d := domain.RealisasiRencanaKerja{
		ID:                      1,
		IndikatorRencanaKerjaID: 2,
		Tahun:                   2026,
		Bulan:                   0,
		Triwulan:                2,
		NilaiRealisasi:          75.5,
		RealisasiAnggaran:       50000,
		Keterangan:              "update triwulan 2",
		DiinputOleh:             9,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	gormModel := ToGormRealisasiRencanaKerja(d)
	if gormModel.Bulan != nil {
		t.Fatal("expected bulan to be nil when domain bulan <= 0")
	}
	if gormModel.Triwulan == nil || *gormModel.Triwulan != 2 {
		t.Fatal("expected triwulan pointer to be populated")
	}

	got := ToDomainRealisasiRencanaKerja(database.RealisasiRencanaKerja{
		ID:                      gormModel.ID,
		IndikatorRencanaKerjaID: gormModel.IndikatorRencanaKerjaID,
		Tahun:                   gormModel.Tahun,
		Bulan:                   gormModel.Bulan,
		Triwulan:                gormModel.Triwulan,
		NilaiRealisasi:          gormModel.NilaiRealisasi,
		RealisasiAnggaran:       gormModel.RealisasiAnggaran,
		Keterangan:              gormModel.Keterangan,
		DiinputOleh:             gormModel.DiinputOleh,
		CreatedAt:               gormModel.CreatedAt,
		UpdatedAt:               gormModel.UpdatedAt,
	})

	if got.Bulan != 0 || got.Triwulan != 2 {
		t.Fatalf("period mapping mismatch: got bulan=%d triwulan=%d", got.Bulan, got.Triwulan)
	}
}
