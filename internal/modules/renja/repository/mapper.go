package repository

import (
	"time"

	"e-plan-ai/internal/modules/renja/domain"
	"e-plan-ai/internal/shared/database"
)

func ToDomainRencanaKerja(m database.RencanaKerja) domain.RencanaKerja {
	triwulan := int8(0)
	if m.Triwulan != nil {
		triwulan = *m.Triwulan
	}

	approvedBy := int64(0)
	if m.DisetujuiOleh != nil {
		approvedBy = int64(*m.DisetujuiOleh)
	}

	approvedAt := time.Time{}
	if m.TanggalPersetujuan != nil {
		approvedAt = *m.TanggalPersetujuan
	}

	rejectedReason := ""
	if m.Status == string(domain.StatusRejected) {
		rejectedReason = m.Catatan
	}

	return domain.RencanaKerja{
		ID:                     int64(m.ID),
		IndikatorSubKegiatanID: int64(m.IndikatorSubKegiatanID),
		Code:                   m.Kode,
		Name:                   m.Nama,
		Tahun:                  m.Tahun,
		Triwulan:               triwulan,
		UnitPengusulID:         int64(m.UnitPengusulID),
		Status:                 domain.Status(m.Status),
		Notes:                  m.Catatan,
		RejectedReason:         rejectedReason,
		CreatedBy:              int64(m.DibuatOleh),
		UpdatedBy:              int64(m.DibuatOleh),
		ApprovedBy:             approvedBy,
		ApprovedAt:             approvedAt,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}

func ToGormRencanaKerja(d domain.RencanaKerja) database.RencanaKerja {
	catatan := d.Notes
	if d.RejectedReason != "" {
		catatan = d.RejectedReason
	}

	return database.RencanaKerja{
		ID:                     uint64(d.ID),
		IndikatorSubKegiatanID: uint64(d.IndikatorSubKegiatanID),
		Kode:                   d.Code,
		Nama:                   d.Name,
		Tahun:                  d.Tahun,
		Triwulan:               nullableInt8(d.Triwulan),
		UnitPengusulID:         uint64(d.UnitPengusulID),
		Status:                 string(d.Status),
		Catatan:                catatan,
		DibuatOleh:             uint64(d.CreatedBy),
		DisetujuiOleh:          nullableUint64(d.ApprovedBy),
		TanggalPersetujuan:     nullableTime(d.ApprovedAt),
		CreatedAt:              d.CreatedAt,
		UpdatedAt:              d.UpdatedAt,
	}
}

func ToDomainIndikatorRencanaKerja(m database.IndikatorRencanaKerja) domain.IndikatorRencanaKerja {
	return domain.IndikatorRencanaKerja{
		ID:              int64(m.ID),
		RencanaKerjaID:  int64(m.RencanaKerjaID),
		Code:            m.Kode,
		Name:            m.Nama,
		Unit:            m.Satuan,
		TargetTahunan:   m.TargetTahunan,
		AnggaranTahunan: m.AnggaranTahunan,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func ToGormIndikatorRencanaKerja(d domain.IndikatorRencanaKerja) database.IndikatorRencanaKerja {
	return database.IndikatorRencanaKerja{
		ID:              uint64(d.ID),
		RencanaKerjaID:  uint64(d.RencanaKerjaID),
		Kode:            d.Code,
		Nama:            d.Name,
		Satuan:          d.Unit,
		TargetTahunan:   d.TargetTahunan,
		AnggaranTahunan: d.AnggaranTahunan,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func ToDomainRealisasiRencanaKerja(m database.RealisasiRencanaKerja) domain.RealisasiRencanaKerja {
	bulan := int8(0)
	if m.Bulan != nil {
		bulan = *m.Bulan
	}

	triwulan := int8(0)
	if m.Triwulan != nil {
		triwulan = *m.Triwulan
	}

	return domain.RealisasiRencanaKerja{
		ID:                      int64(m.ID),
		IndikatorRencanaKerjaID: int64(m.IndikatorRencanaKerjaID),
		Tahun:                   m.Tahun,
		Bulan:                   bulan,
		Triwulan:                triwulan,
		NilaiRealisasi:          m.NilaiRealisasi,
		RealisasiAnggaran:       m.RealisasiAnggaran,
		Keterangan:              m.Keterangan,
		DiinputOleh:             int64(m.DiinputOleh),
		CreatedAt:               m.CreatedAt,
		UpdatedAt:               m.UpdatedAt,
	}
}

func ToGormRealisasiRencanaKerja(d domain.RealisasiRencanaKerja) database.RealisasiRencanaKerja {
	return database.RealisasiRencanaKerja{
		ID:                      uint64(d.ID),
		IndikatorRencanaKerjaID: uint64(d.IndikatorRencanaKerjaID),
		Tahun:                   d.Tahun,
		Bulan:                   nullableInt8(d.Bulan),
		Triwulan:                nullableInt8(d.Triwulan),
		NilaiRealisasi:          d.NilaiRealisasi,
		RealisasiAnggaran:       d.RealisasiAnggaran,
		Keterangan:              d.Keterangan,
		DiinputOleh:             uint64(d.DiinputOleh),
		CreatedAt:               d.CreatedAt,
		UpdatedAt:               d.UpdatedAt,
	}
}

func nullableInt8(v int8) *int8 {
	if v <= 0 {
		return nil
	}
	vv := v
	return &vv
}

func nullableUint64(v int64) *uint64 {
	if v <= 0 {
		return nil
	}
	vv := uint64(v)
	return &vv
}

func nullableTime(v time.Time) *time.Time {
	if v.IsZero() {
		return nil
	}
	vv := v
	return &vv
}
