package repository

import (
	"context"
	"e-plan-ai/internal/modules/dokumen_pdf/model"

	"gorm.io/gorm"
)

type DokumenPDFRepository struct {
    db *gorm.DB
}

func NewDokumenPDFRepository(db *gorm.DB) *DokumenPDFRepository {
    return &DokumenPDFRepository{db: db}
}

func (r *DokumenPDFRepository) Create(ctx context.Context, doc *model.DokumenPDF) error {
    return r.db.WithContext(ctx).Create(doc).Error
}

func (r *DokumenPDFRepository) GetByID(ctx context.Context, id uint) (*model.DokumenPDF, error) {
    var doc model.DokumenPDF
    err := r.db.WithContext(ctx).First(&doc, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &doc, nil
}

func (r *DokumenPDFRepository) List(ctx context.Context) ([]model.DokumenPDF, error) {
    var docs []model.DokumenPDF
    err := r.db.WithContext(ctx).Find(&docs).Error
    return docs, err
}

func (r *DokumenPDFRepository) Update(ctx context.Context, doc *model.DokumenPDF) error {
    return r.db.WithContext(ctx).Save(doc).Error
}

func (r *DokumenPDFRepository) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&model.DokumenPDF{}, "id = ?", id).Error
}

func (r *DokumenPDFRepository) FindLatest(ctx context.Context, dokumen *model.DokumenPDF) error {
    return r.db.WithContext(ctx).Order("tahun DESC").First(dokumen).Error
}

