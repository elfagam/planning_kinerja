package usecase

import (
	"e-plan-ai/internal/shared/database"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SyncTargetDanRealisasi synchronizes aggregated data from realisasi_rencana_kerja to target_dan_realisasi.
func SyncTargetDanRealisasi(db *gorm.DB, rkID uint64, tahun int16, triwulan int8) error {
	if rkID == 0 || tahun == 0 || triwulan < 1 || triwulan > 4 {
		return nil
	}

	// 1. Sum up realizations from raw table for the specific quarter
	var stats struct {
		TotalNilai     float64 `gorm:"column:total_nilai"`
		TotalAnggaran float64 `gorm:"column:total_anggaran"`
	}

	fallbackTriwulanExpr := "CASE WHEN triwulan IS NOT NULL THEN triwulan WHEN bulan BETWEEN 1 AND 3 THEN 1 WHEN bulan BETWEEN 4 AND 6 THEN 2 WHEN bulan BETWEEN 7 AND 9 THEN 3 WHEN bulan BETWEEN 10 AND 12 THEN 4 ELSE 0 END"

	err := db.Table("realisasi_rencana_kerja").
		Select("SUM(nilai_realisasi) as total_nilai, SUM(realisasi_anggaran) as total_anggaran").
		Where("rencana_kerja_id = ? AND tahun = ? AND ("+fallbackTriwulanExpr+") = ?", rkID, tahun, triwulan).
		Scan(&stats).Error
	if err != nil {
		return err
	}

	// 2. Get Target from Rencana Kerja (Annual Target)
	var rk database.RencanaKerja
	if err := db.First(&rk, rkID).Error; err != nil {
		return err
	}

	// 3. Get Target Anggaran from Indikator Rencana Kerja (Annual Budget)
	var targetAnggaranTahunan float64
	db.Table("indikator_rencana_kerja").
		Where("rencana_kerja_id = ?", rkID).
		Select("COALESCE(SUM(anggaran_tahunan), 0)").
		Scan(&targetAnggaranTahunan)

	// Benchmarks for the quarter (assuming equal distribution for simplicity in aggregate)
	targetNilaiTriwulan := rk.Target / 4
	targetAnggaranTriwulan := targetAnggaranTahunan / 4

	// Calculate percentages
	capaianPersen := 0.0
	if targetNilaiTriwulan > 0 {
		capaianPersen = (stats.TotalNilai / targetNilaiTriwulan) * 100
	}
	capaianAnggaran := 0.0
	if targetAnggaranTriwulan > 0 {
		capaianAnggaran = (stats.TotalAnggaran / targetAnggaranTriwulan) * 100
	}

	// Determine status based on performance
	status := "OFF_TRACK"
	if capaianPersen >= 100 {
		status = "ON_TRACK"
	} else if capaianPersen >= 80 {
		status = "WARNING"
	}

	now := time.Now()
	item := database.TargetDanRealisasi{
		RencanaKerjaID:    rkID,
		Tahun:             tahun,
		Triwulan:          triwulan,
		TargetNilai:       targetNilaiTriwulan,
		RealisasiNilai:    stats.TotalNilai,
		CapaianPersen:     capaianPersen,
		TargetAnggaran:    targetAnggaranTriwulan,
		RealisasiAnggaran: stats.TotalAnggaran,
		CapaianAnggaran:   capaianAnggaran,
		Status:            status,
		UpdatedAt:         now,
	}

	// 4. Upsert into TargetDanRealisasi
	// uq_target_realisasi_periode is Unique Index on (rencana_kerja_id, tahun, triwulan)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "rencana_kerja_id"},
			{Name: "tahun"},
			{Name: "triwulan"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"target_nilai",
			"realisasi_nilai",
			// "capaian_persen" is STORED GENERATED, must be omitted
			"target_anggaran",
			"realisasi_anggaran",
			"capaian_anggaran",
			"status",
			"updated_at",
		}),
	}).Omit("capaian_persen").Create(&item).Error
}
