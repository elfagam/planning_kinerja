package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"e-plan-ai/internal/modules/renja/domain"
)

// GenerateIndikatorCSV generates a CSV export for Indikator Kinerja.
func (s *Service) GenerateIndikatorCSV(ctx context.Context, rencanaKerjaID, unitPengusulID uint) ([]byte, error) {
	data, err := s.repo.(interface {
		GetIndikatorCSVData(context.Context, uint, uint) ([]domain.ExportIndikatorCSVFlatDTO, error)
	}).GetIndikatorCSVData(ctx, rencanaKerjaID, unitPengusulID)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no data found for the given filters")
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	// Top Layout
	top := data[0]
	_ = w.Write([]string{"Program:", top.ProgramNama})
	_ = w.Write([]string{"Kegiatan:", top.KegiatanNama})
	_ = w.Write([]string{"Sub Kegiatan:", top.SubKegiatanNama})
	_ = w.Write([]string{"Unit Pengusul:", top.UnitPengusulKode + " - " + top.UnitPengusulNama})
	_ = w.Write([]string{"Tahun:", fmt.Sprintf("%d", top.RencanaKerjaTahun)})
	w.Write([]string{}) // Empty row

	// Table Headers
	headers := []string{"Kode Rencana", "Nama Rencana", "ID Rekening", "ID Standar Harga", "Kode", "Uraian", "Satuan", "Harga Satuan", "Target", "Anggaran"}
	_ = w.Write(headers)

	// Grouping logic: leave grouped columns blank if same as previous row
	var prev domain.ExportIndikatorCSVFlatDTO
	for i, row := range data {
		var rkKode, rkNama, rek, shID string
		if i == 0 || row.RencanaKerjaNama != prev.RencanaKerjaNama {
			rkKode = row.RencanaKerjaKode
			rkNama = row.RencanaKerjaNama
		}
		if i == 0 || row.StandarHargaIdRekening != prev.StandarHargaIdRekening {
			rek = row.StandarHargaIdRekening
		}
		if i == 0 || row.StandarHargaId != prev.StandarHargaId {
			shID = fmt.Sprintf("%d", row.StandarHargaId)
		}
		w.Write([]string{
			rkKode,
			rkNama,
			rek,
			shID,
			row.IndikatorKode,
			row.IndikatorNama,
			row.Satuan,
			fmt.Sprintf("%.2f", row.HargaSatuan),
			fmt.Sprintf("%.2f", row.TargetTahunan),
			fmt.Sprintf("%.2f", row.AnggaranTahunan),
		})
		prev = row
	}

	w.Write([]string{}) // Empty row

	// Bottom Layout (signatures)
	_ = w.Write([]string{"", "", "", "Mengetahui,"})
	_ = w.Write([]string{"", "", "", top.JabatanPenanggungJawab})
	_ = w.Write([]string{"", "", "", top.NamaPenanggungJawab})
	_ = w.Write([]string{"", "", "", top.NipPenanggungJawab})

	w.Flush()
	return buf.Bytes(), nil
}

// GenerateRencanaKerjaCSV generates a specific CSV export for Rencana Kerja with grouping by rekening.
func (s *Service) GenerateRencanaKerjaCSV(ctx context.Context, subKegiatanID uint) ([]byte, error) {
	data, err := s.repo.GetRencanaKerjaCSVData(ctx, subKegiatanID)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("data tidak ditemukan untuk sub kegiatan yang dipilih")
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	// 1. Header Dokumen (Bagian Atas)
	top := data[0]
	_ = w.Write([]string{"Program:", top.ProgramNama})
	_ = w.Write([]string{"Kegiatan:", top.KegiatanNama})
	_ = w.Write([]string{"Sub Kegiatan:", top.SubKegiatanNama})
	_ = w.Write([]string{"Tahun:", fmt.Sprintf("%d", top.RencanaKerjaTahun)})
	_ = w.Write([]string{}) // Satu baris kosong

	// 2. Table Header (Body)
	headers := []string{
		"ID Rekening",
		"ID Standar Harga",
		"Kode Rincian",
		"Nama Rincian",
		"Target",
		"Satuan",
		"Harga Satuan",
		"Anggaran",
		"Kode Rencana",
		"Nama Rencana",
		"Unit Pengusul",
	}
	_ = w.Write(headers)

	// 3. Body Data with Grouping & Sub-Header
	var currentRekening string
	var totalAnggaran float64
	for _, row := range data {
		totalAnggaran += row.AnggaranTahunan
		// Logika Sub-Header: Tampilkan nama/kode rekening jika berubah
		if row.StandarHargaIdRekening != currentRekening {
			currentRekening = row.StandarHargaIdRekening
			rekLabel := currentRekening
			if rekLabel == "" {
				rekLabel = "(Tanpa Rekening)"
			}
			// Baris pemisah kelompok (Sub-Header)
			_ = w.Write([]string{rekLabel, "", "", "KELOMPOK REKENING: " + rekLabel})
		}

		// Row Data
		w.Write([]string{
			row.StandarHargaIdRekening,
			fmt.Sprintf("%d", row.StandarHargaId),
			row.IndikatorKode,
			row.IndikatorNama,
			fmt.Sprintf("%.2f", row.TargetTahunan),
			row.Satuan,
			fmt.Sprintf("%.2f", row.HargaSatuan),
			fmt.Sprintf("%.2f", row.AnggaranTahunan),
			row.RencanaKerjaKode,
			row.RencanaKerjaNama,
			row.UnitPengusulNama,
		})
	}

	// Baris Total Anggaran
	_ = w.Write([]string{"", "", "", "", "", "", "TOTAL ANGGARAN:", fmt.Sprintf("%.2f", totalAnggaran)})

	_ = w.Write([]string{}) // Satu baris kosong di akhir tabel

	// 4. Footer Dokumen (Bagian Bawah)
	today := time.Now().Format("02 January 2006")
	_ = w.Write([]string{"Tanggal:", today})
	_ = w.Write([]string{"Yang Menandatangani,"})
	for i := 0; i < 5; i++ {
		_ = w.Write([]string{""})
	}
	_ = w.Write([]string{".................................."})

	w.Flush()
	return buf.Bytes(), nil
}
