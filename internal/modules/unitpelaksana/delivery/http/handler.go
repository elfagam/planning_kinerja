package http

import (
	"errors"
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
	Kode       string `json:"kode"`
	Nama       string `json:"nama"`
	Keterangan string `json:"keterangan"`
	Aktif      *bool  `json:"aktif"`
}

func NewHandler(cfg config.Config) *Handler {
	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("unit_pelaksana handler unavailable: %v", err)
		return &Handler{ready: false, reason: "database connection unavailable"}
	}

	return &Handler{db: db, ready: true}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/unit-pelaksana")
	group.GET("", h.List)
	group.GET("/:id", h.Get)
	group.POST("", h.Create)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)

	// Alias endpoint for underscore naming convention.
	v1.GET("/unit_pelaksana", h.List)
	v1.POST("/unit_pelaksana", h.Create)
	v1.PUT("/unit_pelaksana/:id", h.Update)
	v1.DELETE("/unit_pelaksana/:id", h.Delete)
}

func (h *Handler) List(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	q := strings.TrimSpace(c.Query("q"))
	query := h.db.WithContext(c.Request.Context()).Model(&database.UnitPelaksana{})
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("kode LIKE ? OR nama LIKE ?", like, like)
	}

	var items []database.UnitPelaksana
	if err := query.Order("id DESC").Find(&items).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list unit_pelaksana")
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

	var item database.UnitPelaksana
	err := h.db.WithContext(c.Request.Context()).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "unit_pelaksana not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get unit_pelaksana")
		return
	}

	response.Success(c, item)
}

func (h *Handler) Create(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if strings.TrimSpace(req.Kode) == "" || strings.TrimSpace(req.Nama) == "" {
		response.Error(c, http.StatusBadRequest, "kode and nama are required")
		return
	}

	aktif := true
	if req.Aktif != nil {
		aktif = *req.Aktif
	}

	item := database.UnitPelaksana{
		Kode:       strings.TrimSpace(req.Kode),
		Nama:       strings.TrimSpace(req.Nama),
		Keterangan: strings.TrimSpace(req.Keterangan),
		Aktif:      aktif,
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to create unit_pelaksana")
		return
	}

	response.Success(c, item)
}

func (h *Handler) Update(c *gin.Context) {
	if !h.ensureReady(c) {
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
	if strings.TrimSpace(req.Kode) == "" || strings.TrimSpace(req.Nama) == "" {
		response.Error(c, http.StatusBadRequest, "kode and nama are required")
		return
	}

	var item database.UnitPelaksana
	err := h.db.WithContext(c.Request.Context()).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "unit_pelaksana not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get unit_pelaksana")
		return
	}

	item.Kode = strings.TrimSpace(req.Kode)
	item.Nama = strings.TrimSpace(req.Nama)
	item.Keterangan = strings.TrimSpace(req.Keterangan)
	if req.Aktif != nil {
		item.Aktif = *req.Aktif
	}

	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to update unit_pelaksana")
		return
	}

	response.Success(c, item)
}

func (h *Handler) Delete(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	res := h.db.WithContext(c.Request.Context()).Delete(&database.UnitPelaksana{}, "id = ?", id)
	if res.Error != nil {
		response.Error(c, http.StatusInternalServerError, "failed to delete unit_pelaksana")
		return
	}
	if res.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "unit_pelaksana not found")
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

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
