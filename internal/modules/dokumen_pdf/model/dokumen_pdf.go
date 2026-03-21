package model

import "time"


type DokumenPDF struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Tahun     int       `json:"tahun"`
    Nama      string    `json:"nama"`
    FilePath  string    `json:"file_path"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (DokumenPDF) TableName() string {
	return "dokumen_pdf"
}