package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"e-plan-ai/internal/modules/renja/domain"
	"e-plan-ai/internal/shared/database"

	"gorm.io/gorm"
)

type RenjaGormRepository struct {
	db            *gorm.DB
	auditTableMux sync.Once
}

// GetIndikatorCSVData fetches flat indikator kinerja data for CSV export.
func (r *RenjaGormRepository) GetIndikatorCSVData(ctx context.Context, rencanaKerjaID, unitPengusulID uint) ([]domain.ExportIndikatorCSVFlatDTO, error) {
	var results []domain.ExportIndikatorCSVFlatDTO
	err := r.dbFromContext(ctx).
		Table("rencana_kerja").
		Select(`
			program.nama as program_nama,
			kegiatan.nama as kegiatan_nama,
			sub_kegiatan.nama as sub_kegiatan_nama,
			unit_pengusul.kode as unit_pengusul_kode,
			unit_pengusul.nama as unit_pengusul_nama,
			rencana_kerja.tahun as rencana_kerja_tahun,
			rencana_kerja.kode as rencana_kerja_kode,
			rencana_kerja.nama as rencana_kerja_nama,
			tb_standar_harga.id_rekening as standar_harga_id_rekening,
			tb_standar_harga.id as standar_harga_id,
			indikator_rencana_kerja.kode as indikator_kode,
			indikator_rencana_kerja.nama as indikator_nama,
			indikator_rencana_kerja.harga_satuan as harga_satuan,
			indikator_rencana_kerja.satuan as satuan,
			indikator_rencana_kerja.target_tahunan as target_tahunan,
			indikator_rencana_kerja.anggaran_tahunan as anggaran_tahunan,
			unit_pengusul.jabatan_penanggungjawab as jabatan_penanggung_jawab,
			unit_pengusul.nama_penanggungjawab as nama_penanggung_jawab,
			unit_pengusul.nip_penanggungjawab as nip_penanggung_jawab
		`).
		Joins("JOIN unit_pengusul ON unit_pengusul.id = rencana_kerja.unit_pengusul_id").
		Joins("JOIN indikator_rencana_kerja ON indikator_rencana_kerja.rencana_kerja_id = rencana_kerja.id").
		Joins("LEFT JOIN tb_standar_harga ON tb_standar_harga.id = indikator_rencana_kerja.tb_standar_harga_id").
		Joins("LEFT JOIN indikator_sub_kegiatan ON indikator_sub_kegiatan.id = rencana_kerja.indikator_sub_kegiatan_id").
		Joins("LEFT JOIN sub_kegiatan ON sub_kegiatan.id = indikator_sub_kegiatan.sub_kegiatan_id").
		Joins("LEFT JOIN kegiatan ON kegiatan.id = sub_kegiatan.kegiatan_id").
		Joins("LEFT JOIN program ON program.id = kegiatan.program_id").
		Where("rencana_kerja.id = ? AND rencana_kerja.unit_pengusul_id = ?", rencanaKerjaID, unitPengusulID).
		Order("rencana_kerja.nama, tb_standar_harga.id_rekening, tb_standar_harga.id, indikator_rencana_kerja.kode").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetRencanaKerjaCSVData fetches flat rencana kerja data filtered by sub_kegiatan_id.
func (r *RenjaGormRepository) GetRencanaKerjaCSVData(ctx context.Context, subKegiatanID uint) ([]domain.ExportIndikatorCSVFlatDTO, error) {
	var results []domain.ExportIndikatorCSVFlatDTO
	err := r.dbFromContext(ctx).
		Table("rencana_kerja").
		Select(`
			program.nama as program_nama,
			kegiatan.nama as kegiatan_nama,
			sub_kegiatan.nama as sub_kegiatan_nama,
			unit_pengusul.kode as unit_pengusul_kode,
			unit_pengusul.nama as unit_pengusul_nama,
			rencana_kerja.tahun as rencana_kerja_tahun,
			rencana_kerja.kode as rencana_kerja_kode,
			rencana_kerja.nama as rencana_kerja_nama,
			tb_standar_harga.id_rekening as standar_harga_id_rekening,
			tb_standar_harga.id as standar_harga_id,
			indikator_rencana_kerja.kode as indikator_kode,
			indikator_rencana_kerja.nama as indikator_nama,
			indikator_rencana_kerja.harga_satuan as harga_satuan,
			indikator_rencana_kerja.satuan as satuan,
			indikator_rencana_kerja.target_tahunan as target_tahunan,
			indikator_rencana_kerja.anggaran_tahunan as anggaran_tahunan,
			unit_pengusul.jabatan_penanggungjawab as jabatan_penanggung_jawab,
			unit_pengusul.nama_penanggungjawab as nama_penanggung_jawab,
			unit_pengusul.nip_penanggungjawab as nip_penanggung_jawab
		`).
		Joins("JOIN unit_pengusul ON unit_pengusul.id = rencana_kerja.unit_pengusul_id").
		Joins("JOIN indikator_rencana_kerja ON indikator_rencana_kerja.rencana_kerja_id = rencana_kerja.id").
		Joins("LEFT JOIN tb_standar_harga ON tb_standar_harga.id = indikator_rencana_kerja.tb_standar_harga_id").
		Joins("LEFT JOIN indikator_sub_kegiatan ON indikator_sub_kegiatan.id = rencana_kerja.indikator_sub_kegiatan_id").
		Joins("LEFT JOIN sub_kegiatan ON sub_kegiatan.id = indikator_sub_kegiatan.sub_kegiatan_id").
		Joins("LEFT JOIN kegiatan ON kegiatan.id = sub_kegiatan.kegiatan_id").
		Joins("LEFT JOIN program ON program.id = kegiatan.program_id").
		Where("sub_kegiatan.id = ?", subKegiatanID).
		Order("tb_standar_harga.id_rekening, tb_standar_harga.id, rencana_kerja.kode, indikator_rencana_kerja.kode").
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rencana kerja csv data: %w", err)
	}
	return results, nil
}

func NewRenjaGormRepository(db *gorm.DB) *RenjaGormRepository {
	return &RenjaGormRepository{db: db}
}

func (r *RenjaGormRepository) GetByID(ctx context.Context, id int64) (domain.Renja, error) {
	if r == nil || r.db == nil {
		return domain.Renja{}, fmt.Errorf("nil renja repository db")
	}

	if id <= 0 {
		return domain.Renja{}, fmt.Errorf("invalid renja id: %d", id)
	}

	var model database.RencanaKerja
	err := r.dbFromContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return domain.Renja{}, err
	}

	return ToDomainRencanaKerja(model), nil
}

func (r *RenjaGormRepository) Save(ctx context.Context, renja domain.Renja) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil renja repository db")
	}

	if renja.ID <= 0 {
		return fmt.Errorf("invalid renja id: %d", renja.ID)
	}

	model := ToGormRencanaKerja(renja)
	updates := map[string]any{
		"status":              model.Status,
		"catatan":             model.Catatan,
		"disetujui_oleh":      model.DisetujuiOleh,
		"tanggal_persetujuan": model.TanggalPersetujuan,
		"updated_at":          time.Now(),
	}

	res := r.dbFromContext(ctx).
		Model(&database.RencanaKerja{}).
		Where("id = ?", model.ID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *RenjaGormRepository) AppendAudit(ctx context.Context, actorID int64, action string, resourceID int64, notes string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil renja repository db")
	}

	if err := r.ensureAuditTable(ctx); err != nil {
		return err
	}

	const insertAudit = `
INSERT INTO renja_audit_logs (actor_id, action, resource_id, notes, created_at)
VALUES (?, ?, ?, ?, ?)`

	return r.dbFromContext(ctx).
		Exec(insertAudit, actorID, action, resourceID, notes, time.Now()).
		Error
}

func (r *RenjaGormRepository) ensureAuditTable(ctx context.Context) error {
	var ensureErr error
	r.auditTableMux.Do(func() {
		const ddl = `
CREATE TABLE IF NOT EXISTS renja_audit_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    actor_id BIGINT NOT NULL,
    action VARCHAR(64) NOT NULL,
    resource_id BIGINT NOT NULL,
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_renja_audit_actor (actor_id),
    INDEX idx_renja_audit_resource (resource_id),
    INDEX idx_renja_audit_action (action)
)`
		ensureErr = r.dbFromContext(ctx).Exec(ddl).Error
	})
	return ensureErr
}

func (r *RenjaGormRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if r.db == nil {
		return nil
	}
	if tx := GormTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}

var _ interface {
	GetByID(context.Context, int64) (domain.Renja, error)
	Save(context.Context, domain.Renja) error
	AppendAudit(context.Context, int64, string, int64, string) error
} = (*RenjaGormRepository)(nil)
