package database

import "time"

type UnitPengusul struct {
	ID                     uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Kode                   string    `json:"kode" gorm:"size:30;not null;uniqueIndex:uq_unit_pengusul_kode"`
	Nama                   string    `json:"nama" gorm:"size:150;not null"`
	NamaPenanggungjawab    string    `json:"nama_penanggungjawab" gorm:"size:150"`
	NipPenanggungjawab     string    `json:"nip_penanggungjawab" gorm:"size:30"`
	JabatanPenanggungjawab string    `json:"jabatan_penanggungjawab" gorm:"size:100"`
	Keterangan             string    `json:"keterangan" gorm:"type:text"`
	Aktif                  bool      `json:"aktif" gorm:"not null;default:true"`
	CreatedAt              time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt              time.Time `json:"updated_at" gorm:"not null"`
}

func (UnitPengusul) TableName() string { return "unit_pengusul" }

type UnitPelaksana struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Kode       string    `json:"kode" gorm:"size:30;not null;uniqueIndex:uq_unit_pelaksana_kode"`
	Nama       string    `json:"nama" gorm:"size:150;not null"`
	Keterangan string    `json:"keterangan" gorm:"type:text"`
	Aktif      bool      `json:"aktif" gorm:"not null;default:true"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"not null"`
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
	Visi      *Visi     `gorm:"foreignKey:VisiID;references:ID;constraint:fk_gorm_misi_visi"`
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
	Misi      *Misi     `gorm:"foreignKey:MisiID;references:ID;constraint:fk_gorm_tujuan_misi"`
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
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Kode      string    `gorm:"size:40;not null;uniqueIndex:uq_program_kode"`
	Nama      string    `gorm:"size:255;not null"`
	Deskripsi string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Program) TableName() string { return "program" }

type IndikatorProgram struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	SasaranID uint64    `gorm:"not null"`
	ProgramID uint64    `gorm:"not null;index:idx_indikator_program_program_id"`
	Program   *Program  `gorm:"foreignKey:ProgramID;references:ID;constraint:fk_gorm_indikator_program_program"`
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
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	ProgramID uint64    `gorm:"not null;index:idx_kegiatan_program_id"`
	Program   *Program  `gorm:"foreignKey:ProgramID;references:ID;constraint:fk_gorm_kegiatan_program"`
	Kode      string    `gorm:"size:40;not null;uniqueIndex:uq_kegiatan_kode"`
	Nama      string    `gorm:"size:255;not null"`
	Deskripsi string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Kegiatan) TableName() string { return "kegiatan" }

type IndikatorKegiatan struct {
	ID                 uint64            `gorm:"primaryKey;autoIncrement"`
	IndikatorProgramID *uint64           `gorm:"index:idx_indikator_kegiatan_program"`
	IndikatorProgram   *IndikatorProgram `gorm:"foreignKey:IndikatorProgramID"`
	KegiatanID         uint64            `gorm:"not null;index:idx_indikator_kegiatan_kegiatan_id"`
	Kegiatan           *Kegiatan         `gorm:"foreignKey:KegiatanID;constraint:fk_gorm_indikator_kegiatan_kegiatan"`
	Kode               string            `gorm:"size:40;not null;uniqueIndex:uq_indikator_kegiatan_kode"`
	Nama               string            `gorm:"size:255;not null"`
	Formula            string            `gorm:"type:text"`
	Satuan             string            `gorm:"size:60"`
	Baseline           float64           `gorm:"type:decimal(18,2);not null;default:0"`
	CreatedAt          time.Time         `gorm:"not null"`
	UpdatedAt          time.Time         `gorm:"not null"`
}

func (IndikatorKegiatan) TableName() string { return "indikator_kegiatan" }

type SubKegiatan struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	KegiatanID uint64    `json:"kegiatan_id" gorm:"not null;index:idx_sub_kegiatan_kegiatan_id"`
	Kegiatan   *Kegiatan `json:"kegiatan" gorm:"foreignKey:KegiatanID;references:ID;constraint:fk_gorm_sub_kegiatan_kegiatan"`
	Kode       string    `json:"kode" gorm:"size:40;not null;uniqueIndex:uq_sub_kegiatan_kode"`
	Nama       string    `json:"nama" gorm:"size:255;not null"`
	Deskripsi  string    `json:"deskripsi" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"not null"`
}

func (SubKegiatan) TableName() string { return "sub_kegiatan" }

type PaguSubKegiatan struct {
	ID                  uint64       `gorm:"primaryKey;autoIncrement"`
	SubKegiatanID       uint64       `gorm:"not null;index:idx_pagu_sub_kegiatan_sub_kegiatan"`
	SubKegiatan         *SubKegiatan `gorm:"foreignKey:SubKegiatanID"`
	Tahun               *int16       `gorm:"type:year"`
	PaguTahunSebelumnya float64      `gorm:"type:decimal(18,2);not null;default:0"`
	PaguTahunIni        float64      `gorm:"type:decimal(18,2);not null;default:0"`
}

func (PaguSubKegiatan) TableName() string { return "pagu_sub_kegiatan" }

type IndikatorSubKegiatan struct {
	ID                  uint64             `json:"id" gorm:"primaryKey;autoIncrement"`
	IndikatorKegiatanID uint64             `json:"indikator_kegiatan_id" gorm:"not null;index:idx_indikator_sub_kegiatan_indikator_kegiatan"`
	IndikatorKegiatan   *IndikatorKegiatan `json:"indikator_kegiatan" gorm:"foreignKey:IndikatorKegiatanID;constraint:fk_gorm_indikator_sub_kegiatan_ind_keg"`
	SubKegiatanID       uint64             `json:"sub_kegiatan_id" gorm:"not null;index:idx_indikator_sub_kegiatan_sub_kegiatan"`
	SubKegiatan         *SubKegiatan       `json:"sub_kegiatan" gorm:"foreignKey:SubKegiatanID;constraint:fk_gorm_indikator_sub_kegiatan_sub_keg"`
	Kode                string             `json:"kode" gorm:"size:40;not null;uniqueIndex:uq_indikator_sub_kegiatan_kode"`
	Nama                string             `json:"nama" gorm:"size:255;not null"`
	Formula             string             `json:"formula" gorm:"type:text"`
	Satuan              string             `json:"satuan" gorm:"size:60"`
	Baseline            float64            `json:"baseline" gorm:"type:decimal(18,2);not null;default:0"`
	AnggaranN1          float64            `json:"anggaran_n1" gorm:"column:anggaran_tahun_sebelumnya;type:decimal(18,2);not null;default:0"`
	AnggaranN           float64            `json:"anggaran_n" gorm:"column:anggaran_tahun_ini;type:decimal(18,2);not null;default:0"`
	CreatedAt           time.Time          `json:"created_at" gorm:"not null"`
	UpdatedAt           time.Time          `json:"updated_at" gorm:"not null"`
}

func (IndikatorSubKegiatan) TableName() string { return "indikator_sub_kegiatan" }

type RencanaKerja struct {
	ID                     uint64                `json:"id" gorm:"primaryKey;autoIncrement"`
	IndikatorSubKegiatanID uint64                `json:"indikator_sub_kegiatan_id" gorm:"not null;index:idx_rencana_kerja_indikator_sub_kegiatan"`
	IndikatorSubKegiatan   *IndikatorSubKegiatan `json:"indikator_sub_kegiatan" gorm:"foreignKey:IndikatorSubKegiatanID;constraint:fk_gorm_rencana_kerja_ind_sub_keg"`
	Kode                   string                `json:"kode" gorm:"size:50;not null;uniqueIndex:uq_rencana_kerja_kode"`
	Nama                   string                `json:"nama" gorm:"size:255;not null"`
	Tahun                  int16                 `json:"tahun" gorm:"not null;index:idx_rencana_kerja_tahun;index:idx_rencana_kerja_periode_status,priority:1;index:idx_rencana_kerja_unit_periode,priority:2"`
	Triwulan               *int8                 `json:"triwulan" gorm:"index:idx_rencana_kerja_periode_status,priority:2;index:idx_rencana_kerja_unit_periode,priority:3"`
	Target                 float64               `json:"target" gorm:"type:decimal(18,2);not null;default:0"`
	Satuan                 string                `json:"satuan" gorm:"size:60"`
	UnitPengusulID         uint64                `json:"unit_pengusul_id" gorm:"not null;index:idx_rencana_kerja_unit_periode,priority:1"`
	UnitPengusul           *UnitPengusul         `json:"unit_pengusul" gorm:"foreignKey:UnitPengusulID;constraint:fk_gorm_rencana_kerja_unit_pengusul"`
	Status                 string                `json:"status" gorm:"type:enum('DRAFT','DIAJUKAN','DISETUJUI','DITOLAK');not null;default:'DRAFT';index:idx_rencana_kerja_status;index:idx_rencana_kerja_periode_status,priority:3"`
	Catatan                string                `json:"catatan" gorm:"type:text"`
	DibuatOleh             uint64                `json:"dibuat_oleh" gorm:"not null"`
	DisetujuiOleh          *uint64               `json:"disetujui_oleh" gorm:""`
	TanggalPersetujuan     *time.Time            `json:"tanggal_persetujuan" gorm:""`
	CreatedAt              time.Time             `json:"created_at" gorm:"not null"`
	UpdatedAt              time.Time             `json:"updated_at" gorm:"not null"`
}

func (RencanaKerja) TableName() string { return "rencana_kerja" }

type IndikatorRencanaKerja struct {
	ID               uint64        `json:"id" gorm:"primaryKey;autoIncrement"`
	RencanaKerjaID   uint64        `json:"rencana_kerja_id" gorm:"not null;index:idx_indikator_rk_rk;index:idx_indikator_rk_rk_kode,priority:1"`
	TbStandarHargaID *uint64        `json:"tb_standar_harga_id" gorm:"column:tb_standar_harga_id"`
	StandarHarga     *StandarHarga `json:"standar_harga" gorm:"foreignKey:TbStandarHargaID;constraint:fk_gorm_indikator_rk_tb_sh"`
	Kode             string        `json:"kode" gorm:"size:50;not null;uniqueIndex:uq_indikator_rk_kode;index:idx_indikator_rk_rk_kode,priority:2"`
	Nama            string    `json:"nama" gorm:"size:255;not null"`
	Satuan          string    `json:"satuan" gorm:"size:60"`
	TargetTahunan   float64   `json:"target_tahunan" gorm:"type:decimal(18,2);not null;default:0"`
	HargaSatuan     float64   `json:"harga_satuan" gorm:"type:decimal(18,2);not null;default:0"`
	AnggaranTahunan float64   `json:"anggaran_tahunan" gorm:"type:decimal(18,2);not null;default:0"`
	DibuatOleh      uint64    `json:"dibuat_oleh" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"not null"`
}

func (IndikatorRencanaKerja) TableName() string { return "indikator_rencana_kerja" }

type RealisasiRencanaKerja struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement"`
	RencanaKerjaID    uint64    `gorm:"not null;column:rencana_kerja_id;index:idx_realisasi_rk_tahun;uniqueIndex:uq_realisasi_rk_periode,priority:1"`
	Tahun             int16     `gorm:"not null;index:idx_realisasi_rk_tahun;index:idx_realisasi_rk_periode,priority:1;index:idx_realisasi_rk_input_user,priority:2;uniqueIndex:uq_realisasi_rk_periode,priority:2"`
	Bulan             *int8     `gorm:"index:idx_realisasi_rk_periode,priority:3;uniqueIndex:uq_realisasi_rk_periode,priority:3"`
	Triwulan          *int8     `gorm:"index:idx_realisasi_rk_periode,priority:2;index:idx_realisasi_rk_input_user,priority:3;uniqueIndex:uq_realisasi_rk_periode,priority:4"`
	NilaiRealisasi    float64   `gorm:"type:decimal(18,2);not null;default:0"`
	RealisasiAnggaran float64   `gorm:"type:decimal(18,2);not null;default:0"`
	Keterangan        string    `gorm:"type:text"`
	DiinputOleh       uint64    `gorm:"not null;index:idx_realisasi_rk_input_user,priority:1"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`
}

func (RealisasiRencanaKerja) TableName() string { return "realisasi_rencana_kerja" }

type TargetDanRealisasi struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement"`
	RencanaKerjaID    uint64     `gorm:"not null;column:rencana_kerja_id;uniqueIndex:uq_target_realisasi_periode,priority:1"`
	RencanaKerja      *RencanaKerja `gorm:"foreignKey:RencanaKerjaID;constraint:fk_target_realisasi_rk"`
	Tahun             int16      `gorm:"not null;index:idx_target_realisasi_periode_status,priority:1;index:idx_target_realisasi_verifikator_periode,priority:2;uniqueIndex:uq_target_realisasi_periode,priority:2"`
	Triwulan          int8       `gorm:"not null;index:idx_target_realisasi_periode_status,priority:2;index:idx_target_realisasi_verifikator_periode,priority:3;uniqueIndex:uq_target_realisasi_periode,priority:3"`
	TargetNilai       float64    `gorm:"type:decimal(18,2);not null;default:0"`
	RealisasiNilai    float64    `gorm:"type:decimal(18,2);not null;default:0"`
	CapaianPersen     float64    `gorm:"type:decimal(8,2)"`
	TargetAnggaran    float64    `gorm:"type:decimal(18,2);not null;default:0"`
	RealisasiAnggaran float64    `gorm:"type:decimal(18,2);not null;default:0"`
	CapaianAnggaran   float64    `gorm:"type:decimal(8,2)"`
	Status            string     `gorm:"type:enum('ON_TRACK','WARNING','OFF_TRACK');not null;default:'ON_TRACK';index:idx_target_realisasi_status;index:idx_target_realisasi_periode_status,priority:3"`
	DiverifikasiOleh  *uint64    `gorm:"index:idx_target_realisasi_verifikator_periode,priority:1"`
	TanggalVerifikasi *time.Time `gorm:""`
	Catatan           string     `gorm:"type:text"`
	CreatedAt         time.Time  `gorm:"not null"`
	UpdatedAt         time.Time  `gorm:"not null"`
}

func (TargetDanRealisasi) TableName() string { return "target_dan_realisasi" }

type Informasi struct {
	ID                        uint64    `gorm:"primaryKey;autoIncrement" json:"id" form:"id"`
	Informasi                 string    `gorm:"type:text;not null" json:"informasi" form:"informasi"`
	Tahun                     int       `gorm:"not null;index:idx_informasi_tahun" json:"tahun" form:"tahun"`
	PilihanRouteHalamanTujuan string    `gorm:"column:pilihan_route_halaman_tujuan;size:120;not null" json:"pilihan_route_halaman_tujuan" form:"pilihan_route_halaman_tujuan"`
	TanggalPembuatan          time.Time `gorm:"column:tanggal_pembuatan;not null" json:"tanggal_pembuatan" form:"tanggal_pembuatan"`
	TanggalUbah               time.Time `gorm:"column:tanggal_ubah;not null" json:"tanggal_ubah" form:"tanggal_ubah"`
}

func (Informasi) TableName() string { return "informasi" }

type StandarHarga struct {
	ID           uint64   `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	JenisStandar *string  `json:"jenis_standar" gorm:"column:jenis_standar"`
	UraianBarang *string  `json:"uraian_barang" gorm:"column:uraian_barang"`
	Spesifikasi  *string  `json:"spesifikasi" gorm:"column:spesifikasi"`
	Satuan       *string  `json:"satuan" gorm:"column:satuan"`
	HargaSatuan  *float64 `json:"harga_satuan" gorm:"column:harga_satuan"`
	IdRekening   *string  `json:"id_rekening" gorm:"column:id_rekening"`
}

func (StandarHarga) TableName() string { return "tb_standar_harga" }
