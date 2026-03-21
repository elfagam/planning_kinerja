package database

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrateAll migrates all planning-system models in dependency-safe order.
func AutoMigrateAll(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("nil gorm db")
	}

	if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		&UnitPengusul{},
		&UnitPelaksana{},
		&User{},
		&Visi{},
		&Misi{},
		&Tujuan{},
		&IndikatorTujuan{},
		&Sasaran{},
		&IndikatorSasaran{},
		&Program{},
		&IndikatorProgram{},
		&Kegiatan{},
		&IndikatorKegiatan{},
		&SubKegiatan{},
		&PaguSubKegiatan{},
		&IndikatorSubKegiatan{},
		&RencanaKerja{},
		&IndikatorRencanaKerja{},
		&RealisasiRencanaKerja{},
		&TargetDanRealisasi{},
		&Informasi{},
	); err != nil {
		return fmt.Errorf("gorm automigrate all models: %w", err)
	}

	return nil
}
