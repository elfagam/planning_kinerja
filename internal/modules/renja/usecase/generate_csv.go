package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

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
