package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	ready  bool
	reason string
}

type upsertRequest struct {
	Kode         string `json:"kode"`
	Nama         string `json:"nama"`
	Deskripsi    string `json:"deskripsi"`
	TahunMulai   int16  `json:"tahun_mulai"`
	TahunSelesai int16  `json:"tahun_selesai"`
	Aktif        *bool  `json:"aktif"`
}

func NewHandler(cfg *config.Config) *Handler {
	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("visi handler unavailable: %v", err)
		return &Handler{ready: false, reason: "database connection unavailable"}
	}

	return &Handler{db: db, ready: true}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/visi")
	group.GET("", h.List)
	group.GET("/:id", h.Get)
	group.POST("", h.Create)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}

func (h *Handler) List(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	q := strings.TrimSpace(c.Query("q"))
	query := h.db.WithContext(c.Request.Context()).Model(&database.Visi{})
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("kode LIKE ? OR nama LIKE ?", like, like)
	}

	var items []database.Visi
	if err := query.Order("id DESC").Find(&items).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list visi")
		return
	}

	response.Success(c, gin.H{"items": items, "total": len(items)})
}

func (h *Handler) Get(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	var item database.Visi
	err := h.db.WithContext(c.Request.Context()).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "visi not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get visi")
		return
	}

	response.Success(c, item)
}

func (h *Handler) Create(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}
	if !allowedToMutateVisi(c) {
		return
	}

	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if !validateUpsertRequest(c, req) {
		return
	}

	aktif := true
	if req.Aktif != nil {
		aktif = *req.Aktif
	}

	item := database.Visi{
		Kode:         strings.TrimSpace(req.Kode),
		Nama:         strings.TrimSpace(req.Nama),
		Deskripsi:    strings.TrimSpace(req.Deskripsi),
		TahunMulai:   req.TahunMulai,
		TahunSelesai: req.TahunSelesai,
		Aktif:        aktif,
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to create visi")
		return
	}

	response.Success(c, item)
}

func (h *Handler) Update(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}
	if !allowedToMutateVisi(c) {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if !validateUpsertRequest(c, req) {
		return
	}

	var item database.Visi
	err := h.db.WithContext(c.Request.Context()).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "visi not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get visi")
		return
	}

	item.Kode = strings.TrimSpace(req.Kode)
	item.Nama = strings.TrimSpace(req.Nama)
	item.Deskripsi = strings.TrimSpace(req.Deskripsi)
	item.TahunMulai = req.TahunMulai
	item.TahunSelesai = req.TahunSelesai
	if req.Aktif != nil {
		item.Aktif = *req.Aktif
	}

	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to update visi")
		return
	}

	response.Success(c, item)
}

func (h *Handler) Delete(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}
	if !allowedToMutateVisi(c) {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	res := h.db.WithContext(c.Request.Context()).Delete(&database.Visi{}, "id = ?", id)
	if res.Error != nil {
		response.Error(c, http.StatusInternalServerError, "failed to delete visi")
		return
	}
	if res.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "visi not found")
		return
	}

	response.Success(c, gin.H{"deleted": id})
}

func (h *Handler) ensureReady(c *gin.Context) bool {
	if h.ready {
		return true
	}
	if h.reason == "" {
		h.reason = "service unavailable"
	}
	response.Error(c, http.StatusServiceUnavailable, h.reason)
	return false
}

// visiMutationAllowedRoles are roles permitted to create/update/delete visi.
var visiMutationAllowedRoles = map[string]bool{
	"ADMIN":     true,
	"OPERATOR":  true,
	"PERENCANA": true,
}

func allowedToMutateVisi(c *gin.Context) bool {
	rawRole, _ := c.Get("auth.role")
	role := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", rawRole)))
	if !visiMutationAllowedRoles[role] {
		response.Error(c, http.StatusForbidden, "anda tidak memiliki hak akses untuk mengubah data visi")
		return false
	}
	return true
}

func validateUpsertRequest(c *gin.Context, req upsertRequest) bool {
	if strings.TrimSpace(req.Kode) == "" || strings.TrimSpace(req.Nama) == "" {
		response.Error(c, http.StatusBadRequest, "kode and nama are required")
		return false
	}
	if req.TahunMulai <= 0 || req.TahunSelesai <= 0 {
		response.Error(c, http.StatusBadRequest, "tahun_mulai and tahun_selesai are required")
		return false
	}
	if req.TahunMulai > req.TahunSelesai {
		response.Error(c, http.StatusBadRequest, "tahun_mulai cannot be greater than tahun_selesai")
		return false
	}
	return true
}

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
