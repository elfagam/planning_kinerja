package domain

import "time"

// Visi represents the top-level strategic vision.
type Visi struct {
	ID           uint64    `json:"id"`
	Kode         string    `json:"kode"`
	Nama         string    `json:"nama"`
	Deskripsi    string    `json:"deskripsi,omitempty"`
	TahunMulai   int16     `json:"tahun_mulai"`
	TahunSelesai int16     `json:"tahun_selesai"`
	Aktif        bool      `json:"aktif"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Misi represents a mission under a vision.
type Misi struct {
	ID        uint64    `json:"id"`
	VisiID    uint64    `json:"visi_id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Deskripsi string    `json:"deskripsi,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tujuan represents a strategic objective under a mission.
type Tujuan struct {
	ID        uint64    `json:"id"`
	MisiID    uint64    `json:"misi_id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Deskripsi string    `json:"deskripsi,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IndikatorTujuan represents KPI/indicator records for a tujuan.
type IndikatorTujuan struct {
	ID        uint64    `json:"id"`
	TujuanID  uint64    `json:"tujuan_id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Formula   string    `json:"formula,omitempty"`
	Satuan    string    `json:"satuan,omitempty"`
	Baseline  float64   `json:"baseline"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Sasaran represents targets under a tujuan.
type Sasaran struct {
	ID        uint64    `json:"id"`
	TujuanID  uint64    `json:"tujuan_id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Deskripsi string    `json:"deskripsi,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IndikatorSasaran represents KPI/indicator records for a sasaran.
type IndikatorSasaran struct {
	ID        uint64    `json:"id"`
	SasaranID uint64    `json:"sasaran_id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Formula   string    `json:"formula,omitempty"`
	Satuan    string    `json:"satuan,omitempty"`
	Baseline  float64   `json:"baseline"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Program represents an executable program under a sasaran.
type Program struct {
	ID             uint64    `json:"id"`
	SasaranID      uint64    `json:"sasaran_id"`
	UnitPengusulID uint64    `json:"unit_pengusul_id"`
	Kode           string    `json:"kode"`
	Nama           string    `json:"nama"`
	Deskripsi      string    `json:"deskripsi,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// IndikatorProgram represents KPI/indicator records for a program.
type IndikatorProgram struct {
	ID        uint64    `json:"id"`
	ProgramID uint64    `json:"program_id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Formula   string    `json:"formula,omitempty"`
	Satuan    string    `json:"satuan,omitempty"`
	Baseline  float64   `json:"baseline"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Kegiatan represents activities under a program.
type Kegiatan struct {
	ID              uint64    `json:"id"`
	ProgramID       uint64    `json:"program_id"`
	UnitPelaksanaID uint64    `json:"unit_pelaksana_id"`
	Kode            string    `json:"kode"`
	Nama            string    `json:"nama"`
	Deskripsi       string    `json:"deskripsi,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// IndikatorKegiatan represents KPI/indicator records for a kegiatan.
type IndikatorKegiatan struct {
	ID         uint64    `json:"id"`
	KegiatanID uint64    `json:"kegiatan_id"`
	Kode       string    `json:"kode"`
	Nama       string    `json:"nama"`
	Formula    string    `json:"formula,omitempty"`
	Satuan     string    `json:"satuan,omitempty"`
	Baseline   float64   `json:"baseline"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SubKegiatan represents sub-activities under a kegiatan.
type SubKegiatan struct {
	ID         uint64    `json:"id"`
	KegiatanID uint64    `json:"kegiatan_id"`
	Kode       string    `json:"kode"`
	Nama       string    `json:"nama"`
	Deskripsi  string    `json:"deskripsi,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// IndikatorSubKegiatan represents KPI/indicator records for a sub kegiatan.
type IndikatorSubKegiatan struct {
	ID            uint64    `json:"id"`
	SubKegiatanID uint64    `json:"sub_kegiatan_id"`
	Kode          string    `json:"kode"`
	Nama          string    `json:"nama"`
	Formula       string    `json:"formula,omitempty"`
	Satuan        string    `json:"satuan,omitempty"`
	Baseline      float64   `json:"baseline"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// HierarchyModule represents one node in the planning hierarchy tree.
type HierarchyModule struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	ParentKey string `json:"parent_key,omitempty"`
}

// HierarchyResponse is the API response payload for planning hierarchy.
type HierarchyResponse struct {
	Modules []HierarchyModule `json:"modules"`
	Status  string            `json:"status"`
}
