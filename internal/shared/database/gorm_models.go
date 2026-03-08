package database

import "time"

type UnitPengusul struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	Kode       string    `gorm:"size:30;not null;uniqueIndex:uq_unit_pengusul_kode"`
	Nama       string    `gorm:"size:150;not null"`
	Keterangan string    `gorm:"type:text"`
	Aktif      bool      `gorm:"not null;default:true"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (UnitPengusul) TableName() string { return "unit_pengusul" }

type UnitPelaksana struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	Kode       string    `gorm:"size:30;not null;uniqueIndex:uq_unit_pelaksana_kode"`
	Nama       string    `gorm:"size:150;not null"`
	Keterangan string    `gorm:"type:text"`
	Aktif      bool      `gorm:"not null;default:true"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (UnitPelaksana) TableName() string { return "unit_pelaksana" }

type User struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	UnitPengusulID  *uint64   `gorm:"index:idx_users_unit_pengusul"`
	UnitPelaksanaID *uint64   `gorm:"index:idx_users_unit_pelaksana"`
	NamaLengkap     string    `gorm:"size:150;not null"`
	Email           string    `gorm:"size:150;not null;uniqueIndex:uq_users_email"`
	PasswordHash    string    `gorm:"size:255;not null"`
	Role            string    `gorm:"type:enum('ADMIN','PERENCANA','VERIFIKATOR','PIMPINAN');not null;default:'PERENCANA'"`
	Aktif           bool      `gorm:"not null;default:true"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (User) TableName() string { return "users" }

type Visi struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Kode         string    `gorm:"size:30;not null;uniqueIndex:uq_visi_kode"`
	Nama         string    `gorm:"size:255;not null"`
	Deskripsi    string    `gorm:"type:text"`
	TahunMulai   int16     `gorm:"not null"`
	TahunSelesai int16     `gorm:"not null"`
	Aktif        bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (Visi) TableName() string { return "visi" }

type Misi struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	VisiID    uint64    `gorm:"not null;index:idx_misi_visi"`
	Kode      string    `gorm:"size:30;not null;uniqueIndex:uq_misi_kode"`
	Nama      string    `gorm:"size:255;not null"`
	Deskripsi string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Misi) TableName() string { return "misi" }

type Tujuan struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	MisiID    uint64    `gorm:"not null;index:idx_tujuan_misi"`
	Kode      string    `gorm:"size:30;not null;uniqueIndex:uq_tujuan_kode"`
	Nama      string    `gorm:"size:255;not null"`
	Deskripsi string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Tujuan) TableName() string { return "tujuan" }

type IndikatorTujuan struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	TujuanID  uint64    `gorm:"not null;index:idx_indikator_tujuan_tujuan"`
	Kode      string    `gorm:"size:40;not null;uniqueIndex:uq_indikator_tujuan_kode"`
	Nama      string    `gorm:"size:255;not null"`
	Formula   string    `gorm:"type:text"`
	Satuan    string    `gorm:"size:60"`
	Baseline  float64   `gorm:"type:decimal(18,2);not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (IndikatorTujuan) TableName() string { return "indikator_tujuan" }

type Sasaran struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	TujuanID  uint64    `gorm:"not null;index:idx_sasaran_tujuan"`
	Kode      string    `gorm:"size:30;not null;uniqueIndex:uq_sasaran_kode"`
	Nama      string    `gorm:"size:255;not null"`
	Deskripsi string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Sasaran) TableName() string { return "sasaran" }

type IndikatorSasaran struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	SasaranID uint64    `gorm:"not null;index:idx_indikator_sasaran_sasaran"`
	Kode      string    `gorm:"size:40;not null;uniqueIndex:uq_indikator_sasaran_kode"`
	Nama      string    `gorm:"size:255;not null"`
	Formula   string    `gorm:"type:text"`
	Satuan    string    `gorm:"size:60"`
	Baseline  float64   `gorm:"type:decimal(18,2);not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (IndikatorSasaran) TableName() string { return "indikator_sasaran" }

type Program struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement"`
	SasaranID      uint64    `gorm:"not null;index:idx_program_sasaran"`
	UnitPengusulID uint64    `gorm:"not null;index:idx_program_unit_pengusul;index:idx_program_unit_sasaran,priority:1"`
	Kode           string    `gorm:"size:40;not null;uniqueIndex:uq_program_kode"`
	Nama           string    `gorm:"size:255;not null"`
	Deskripsi      string    `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (Program) TableName() string { return "program" }

type IndikatorProgram struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	ProgramID uint64    `gorm:"not null;index:idx_indikator_program_program"`
	Kode      string    `gorm:"size:40;not null;uniqueIndex:uq_indikator_program_kode"`
	Nama      string    `gorm:"size:255;not null"`
	Formula   string    `gorm:"type:text"`
	Satuan    string    `gorm:"size:60"`
	Baseline  float64   `gorm:"type:decimal(18,2);not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (IndikatorProgram) TableName() string { return "indikator_program" }

type Kegiatan struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	ProgramID       uint64    `gorm:"not null;index:idx_kegiatan_program"`
	UnitPelaksanaID uint64    `gorm:"not null;index:idx_kegiatan_unit_pelaksana;index:idx_kegiatan_unit_program,priority:1"`
	Kode            string    `gorm:"size:40;not null;uniqueIndex:uq_kegiatan_kode"`
	Nama            string    `gorm:"size:255;not null"`
	Deskripsi       string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (Kegiatan) TableName() string { return "kegiatan" }

type IndikatorKegiatan struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	KegiatanID uint64    `gorm:"not null;index:idx_indikator_kegiatan_kegiatan"`
	Kode       string    `gorm:"size:40;not null;uniqueIndex:uq_indikator_kegiatan_kode"`
	Nama       string    `gorm:"size:255;not null"`
	Formula    string    `gorm:"type:text"`
	Satuan     string    `gorm:"size:60"`
	Baseline   float64   `gorm:"type:decimal(18,2);not null;default:0"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (IndikatorKegiatan) TableName() string { return "indikator_kegiatan" }

type SubKegiatan struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	KegiatanID uint64    `gorm:"not null;index:idx_sub_kegiatan_kegiatan"`
	Kode       string    `gorm:"size:40;not null;uniqueIndex:uq_sub_kegiatan_kode"`
	Nama       string    `gorm:"size:255;not null"`
	Deskripsi  string    `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (SubKegiatan) TableName() string { return "sub_kegiatan" }

type IndikatorSubKegiatan struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	SubKegiatanID uint64    `gorm:"not null;index:idx_indikator_sub_kegiatan_sub_kegiatan"`
	Kode          string    `gorm:"size:40;not null;uniqueIndex:uq_indikator_sub_kegiatan_kode"`
	Nama          string    `gorm:"size:255;not null"`
	Formula       string    `gorm:"type:text"`
	Satuan        string    `gorm:"size:60"`
	Baseline      float64   `gorm:"type:decimal(18,2);not null;default:0"`
	AnggaranN1    float64   `gorm:"column:anggaran_tahun_sebelumnya;type:decimal(18,2);not null;default:0"`
	AnggaranN     float64   `gorm:"column:anggaran_tahun_ini;type:decimal(18,2);not null;default:0"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (IndikatorSubKegiatan) TableName() string { return "indikator_sub_kegiatan" }

type RencanaKerja struct {
	ID                     uint64     `gorm:"primaryKey;autoIncrement"`
	IndikatorSubKegiatanID uint64     `gorm:"not null;index:idx_rencana_kerja_indikator_sub_kegiatan"`
	Kode                   string     `gorm:"size:50;not null;uniqueIndex:uq_rencana_kerja_kode"`
	Nama                   string     `gorm:"size:255;not null"`
	Tahun                  int16      `gorm:"not null;index:idx_rencana_kerja_tahun;index:idx_rencana_kerja_periode_status,priority:1;index:idx_rencana_kerja_unit_periode,priority:2"`
	Triwulan               *int8      `gorm:"index:idx_rencana_kerja_periode_status,priority:2;index:idx_rencana_kerja_unit_periode,priority:3"`
	UnitPengusulID         uint64     `gorm:"not null;index:idx_rencana_kerja_unit_periode,priority:1"`
	Status                 string     `gorm:"type:enum('DRAFT','DIAJUKAN','DISETUJUI','DITOLAK');not null;default:'DRAFT';index:idx_rencana_kerja_status;index:idx_rencana_kerja_periode_status,priority:3"`
	Catatan                string     `gorm:"type:text"`
	DibuatOleh             uint64     `gorm:"not null"`
	DisetujuiOleh          *uint64    `gorm:""`
	TanggalPersetujuan     *time.Time `gorm:""`
	CreatedAt              time.Time  `gorm:"not null"`
	UpdatedAt              time.Time  `gorm:"not null"`
}

func (RencanaKerja) TableName() string { return "rencana_kerja" }

type IndikatorRencanaKerja struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	RencanaKerjaID  uint64    `gorm:"not null;index:idx_indikator_rk_rk;index:idx_indikator_rk_rk_kode,priority:1"`
	Kode            string    `gorm:"size:50;not null;uniqueIndex:uq_indikator_rk_kode;index:idx_indikator_rk_rk_kode,priority:2"`
	Nama            string    `gorm:"size:255;not null"`
	Satuan          string    `gorm:"size:60"`
	TargetTahunan   float64   `gorm:"type:decimal(18,2);not null;default:0"`
	AnggaranTahunan float64   `gorm:"type:decimal(18,2);not null;default:0"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (IndikatorRencanaKerja) TableName() string { return "indikator_rencana_kerja" }

type RealisasiRencanaKerja struct {
	ID                      uint64    `gorm:"primaryKey;autoIncrement"`
	IndikatorRencanaKerjaID uint64    `gorm:"not null;index:idx_realisasi_rk_tahun;uniqueIndex:uq_realisasi_rk_periode,priority:1"`
	Tahun                   int16     `gorm:"not null;index:idx_realisasi_rk_tahun;index:idx_realisasi_rk_periode,priority:1;index:idx_realisasi_rk_input_user,priority:2;uniqueIndex:uq_realisasi_rk_periode,priority:2"`
	Bulan                   *int8     `gorm:"index:idx_realisasi_rk_periode,priority:3;uniqueIndex:uq_realisasi_rk_periode,priority:3"`
	Triwulan                *int8     `gorm:"index:idx_realisasi_rk_periode,priority:2;index:idx_realisasi_rk_input_user,priority:3;uniqueIndex:uq_realisasi_rk_periode,priority:4"`
	NilaiRealisasi          float64   `gorm:"type:decimal(18,2);not null;default:0"`
	RealisasiAnggaran       float64   `gorm:"type:decimal(18,2);not null;default:0"`
	Keterangan              string    `gorm:"type:text"`
	DiinputOleh             uint64    `gorm:"not null;index:idx_realisasi_rk_input_user,priority:1"`
	CreatedAt               time.Time `gorm:"not null"`
	UpdatedAt               time.Time `gorm:"not null"`
}

func (RealisasiRencanaKerja) TableName() string { return "realisasi_rencana_kerja" }

type TargetDanRealisasi struct {
	ID                      uint64     `gorm:"primaryKey;autoIncrement"`
	IndikatorRencanaKerjaID uint64     `gorm:"not null;uniqueIndex:uq_target_realisasi_periode,priority:1"`
	Tahun                   int16      `gorm:"not null;index:idx_target_realisasi_periode_status,priority:1;index:idx_target_realisasi_verifikator_periode,priority:2;uniqueIndex:uq_target_realisasi_periode,priority:2"`
	Triwulan                int8       `gorm:"not null;index:idx_target_realisasi_periode_status,priority:2;index:idx_target_realisasi_verifikator_periode,priority:3;uniqueIndex:uq_target_realisasi_periode,priority:3"`
	TargetNilai             float64    `gorm:"type:decimal(18,2);not null;default:0"`
	RealisasiNilai          float64    `gorm:"type:decimal(18,2);not null;default:0"`
	CapaianPersen           float64    `gorm:"type:decimal(8,2)"`
	Status                  string     `gorm:"type:enum('ON_TRACK','WARNING','OFF_TRACK');not null;default:'ON_TRACK';index:idx_target_realisasi_status;index:idx_target_realisasi_periode_status,priority:3"`
	DiverifikasiOleh        *uint64    `gorm:"index:idx_target_realisasi_verifikator_periode,priority:1"`
	TanggalVerifikasi       *time.Time `gorm:""`
	Catatan                 string     `gorm:"type:text"`
	CreatedAt               time.Time  `gorm:"not null"`
	UpdatedAt               time.Time  `gorm:"not null"`
}

func (TargetDanRealisasi) TableName() string { return "target_dan_realisasi" }
