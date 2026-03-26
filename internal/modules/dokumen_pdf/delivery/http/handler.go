package http

import (
	"e-plan-ai/internal/modules/dokumen_pdf/model"
	"e-plan-ai/internal/modules/dokumen_pdf/repository"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
    repo *repository.DokumenPDFRepository
}

// GetLatestDokumenPDF returns the latest PDF document.
func (h *Handler) GetLatestDokumenPDF(c *gin.Context) {
    var doc model.DokumenPDF
    err := h.repo.FindLatest(c.Request.Context(), &doc)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"success": false, "error": "Belum ada dokumen yang diunggah"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"success": true, "data": doc})
}

func NewHandler(db *gorm.DB) *Handler {
    return &Handler{repo: repository.NewDokumenPDFRepository(db)}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
    group := rg.Group("/dokumen_pdf")
    group.GET("", h.List)
    group.POST("", h.Create)
    group.GET("/:id", h.GetByID)
    group.PUT("/:id", h.Update)
    group.DELETE("/:id", h.Delete)
}

func (h *Handler) List(c *gin.Context) {
    docs, err := h.repo.List(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"items": docs})
}

func (h *Handler) Create(c *gin.Context) {
    tahun := c.PostForm("tahun")
    nama := c.PostForm("nama")
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "file PDF wajib diupload"})
        return
    }

    // Validasi ekstensi
    if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya file PDF yang diizinkan"})
        return
    }

    // Validasi tahun
    tahunInt, err := strconv.Atoi(tahun)
    if tahun == "" || err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "tahun wajib diisi dan harus berupa angka"})
        return
    }

    // Pastikan direktori ada
    uploadDir := "web/assets/pdf/"
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.MkdirAll(uploadDir, 0755)
    }

    // Simpan file PDF ke server dengan nama unik
    uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
    savePath := uploadDir + uniqueFilename
    if err := c.SaveUploadedFile(file, savePath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan file PDF"})
        return
    }

    // Simpan ke database
    doc := model.DokumenPDF{
        Tahun:    tahunInt,
        Nama:     nama,
        FilePath: "/assets/pdf/" + uniqueFilename,
    }
    if err := h.repo.Create(c.Request.Context(), &doc); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, doc)
}

func (h *Handler) GetByID(c *gin.Context) {
    idParam := c.Param("id")
    var id uint
    if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    doc, err := h.repo.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, doc)
}

func (h *Handler) Update(c *gin.Context) {
    idParam := c.Param("id")
    var id uint
    if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    existing, err := h.repo.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "dokumen tidak ditemukan"})
        return
    }

    tahun := c.PostForm("tahun")
    nama := c.PostForm("nama")
    file, _ := c.FormFile("file") // File opsional di update

    if tahun != "" {
        if t, err := strconv.Atoi(tahun); err == nil {
            existing.Tahun = t
        }
    }
    if nama != "" {
        existing.Nama = nama
    }

    // Jika ada file baru yang diupload
    if file != nil {
        // Hapus file lama jika ada
        if existing.FilePath != "" {
            _ = removeFile("web" + existing.FilePath)
        }

        // Simpan file baru
        uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
        savePath := "web/assets/pdf/" + uniqueFilename
        if err := c.SaveUploadedFile(file, savePath); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan file PDF baru"})
            return
        }
        existing.FilePath = "/assets/pdf/" + uniqueFilename
    }

    if err := h.repo.Update(c.Request.Context(), existing); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, existing)
}

func (h *Handler) Delete(c *gin.Context) {
    idParam := c.Param("id")
    var id uint
    if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    // Ambil dokumen untuk dapatkan path file
    doc, err := h.repo.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    // Hapus file PDF jika ada
    if doc.FilePath != "" {
        filePath := "web" + doc.FilePath // FilePath: /assets/pdf/xxx.pdf
        if err := removeFile(filePath); err != nil {
            // Log error, tapi tetap lanjut hapus DB
            fmt.Printf("Gagal hapus file: %v\n", err)
        }
    }
    if err := h.repo.Delete(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.Status(http.StatusNoContent)
}

// Helper hapus file
func removeFile(path string) error {
    return os.Remove(path)
}
