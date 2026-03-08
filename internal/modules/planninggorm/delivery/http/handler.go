package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type resourceConfig struct {
	path         string
	name         string
	requiredKeys []string
	searchFields []string
	newModel     func() any
	newSlice     func() any
}

type Handler struct {
	db        *gorm.DB
	ready     bool
	reason    string
	resources []resourceConfig
}

func NewHandler(cfg config.Config) *Handler {
	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("planninggorm handler unavailable: %v", err)
		return &Handler{ready: false, reason: "database connection unavailable"}
	}

	return &Handler{db: db, ready: true, resources: planningResources()}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	for _, rc := range h.resources {
		cfg := rc
		g := v1.Group("/" + cfg.path)
		g.GET("", h.list(cfg))
		g.GET("/:id", h.get(cfg))
		g.POST("", h.create(cfg))
		g.PUT("/:id", h.update(cfg))
		g.DELETE("/:id", h.delete(cfg))
	}
}

func (h *Handler) list(rc resourceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.ensureReady(c) {
			return
		}

		q := strings.TrimSpace(c.Query("q"))
		query := h.db.WithContext(c.Request.Context()).Model(rc.newModel())
		if q != "" && len(rc.searchFields) > 0 {
			like := "%" + q + "%"
			conds := make([]string, 0, len(rc.searchFields))
			args := make([]any, 0, len(rc.searchFields))
			for _, f := range rc.searchFields {
				conds = append(conds, f+" LIKE ?")
				args = append(args, like)
			}
			query = query.Where(strings.Join(conds, " OR "), args...)
		}

		items := rc.newSlice()
		if err := query.Order("id DESC").Find(items).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to list "+rc.name)
			return
		}

		if rc.path == "program" {
			programs := items.(*[]database.Program)
			responseItems := make([]gin.H, 0, len(*programs))
			for _, p := range *programs {
				responseItems = append(responseItems, programToResponse(p))
			}
			response.Success(c, gin.H{"items": responseItems, "total": len(responseItems)})
			return
		}

		response.Success(c, gin.H{"items": items, "total": reflect.ValueOf(items).Elem().Len()})
	}
}

func (h *Handler) get(rc resourceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.ensureReady(c) {
			return
		}

		id, ok := parseID(c)
		if !ok {
			return
		}

		item := rc.newModel()
		err := h.db.WithContext(c.Request.Context()).First(item, "id = ?", id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, rc.name+" not found")
			return
		}
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to get "+rc.name)
			return
		}

		if rc.path == "program" {
			program := item.(*database.Program)
			response.Success(c, programToResponse(*program))
			return
		}

		response.Success(c, item)
	}
}

func (h *Handler) create(rc resourceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.ensureReady(c) {
			return
		}

		payload := map[string]any{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			response.Error(c, http.StatusBadRequest, "invalid payload")
			return
		}

		if rc.path == "kegiatan" {
			if _, ok := payload["unit_pelaksana_id"]; !ok {
				defaultUnitID, err := h.defaultUnitPelaksanaID(c)
				if err != nil {
					response.Error(c, http.StatusBadRequest, err.Error())
					return
				}
				// Backward-compatible: fill default when client omits unit_pelaksana_id.
				payload["unit_pelaksana_id"] = defaultUnitID
			}
		}

		if !validateRequired(c, payload, rc.requiredKeys) {
			return
		}

		if rc.path == "program" {
			defaultUnitID, err := h.defaultUnitPengusulID(c)
			if err != nil {
				response.Error(c, http.StatusBadRequest, err.Error())
				return
			}
			// Program API no longer accepts unit_pengusul_id from client.
			payload["unit_pengusul_id"] = defaultUnitID
		}

		if rc.path == "kegiatan" {
			defaultUnitID, err := h.defaultUnitPelaksanaID(c)
			if err != nil {
				response.Error(c, http.StatusBadRequest, err.Error())
				return
			}
			// Kegiatan API no longer accepts unit_pelaksana_id from client.
			payload["unit_pelaksana_id"] = defaultUnitID
		}

		now := time.Now()
		if _, ok := payload["created_at"]; !ok {
			payload["created_at"] = now
		}
		if _, ok := payload["updated_at"]; !ok {
			payload["updated_at"] = now
		}

		if err := h.db.WithContext(c.Request.Context()).Model(rc.newModel()).Create(payload).Error; err != nil {
			response.Error(c, http.StatusBadRequest, mapWriteError(rc, "create", err))
			return
		}

		if rc.path == "program" {
			delete(payload, "unit_pengusul_id")
		}
		if rc.path == "kegiatan" {
			delete(payload, "unit_pelaksana_id")
		}

		response.Success(c, payload)
	}
}

func (h *Handler) update(rc resourceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.ensureReady(c) {
			return
		}

		id, ok := parseID(c)
		if !ok {
			return
		}

		payload := map[string]any{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			response.Error(c, http.StatusBadRequest, "invalid payload")
			return
		}
		if !validateRequired(c, payload, rc.requiredKeys) {
			return
		}
		if rc.path == "program" {
			// Program API no longer updates unit_pengusul_id.
			delete(payload, "unit_pengusul_id")
		}
		if rc.path == "kegiatan" {
			// Kegiatan API no longer updates unit_pelaksana_id.
			delete(payload, "unit_pelaksana_id")
		}
		delete(payload, "id")
		delete(payload, "created_at")
		delete(payload, "updated_at")

		res := h.db.WithContext(c.Request.Context()).Model(rc.newModel()).Where("id = ?", id).Updates(payload)
		if res.Error != nil {
			response.Error(c, http.StatusBadRequest, mapWriteError(rc, "update", res.Error))
			return
		}
		if res.RowsAffected == 0 {
			response.Error(c, http.StatusNotFound, rc.name+" not found")
			return
		}

		payload["id"] = id
		response.Success(c, payload)
	}
}

func (h *Handler) delete(rc resourceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.ensureReady(c) {
			return
		}

		id, ok := parseID(c)
		if !ok {
			return
		}

		res := h.db.WithContext(c.Request.Context()).Delete(rc.newModel(), "id = ?", id)
		if res.Error != nil {
			response.Error(c, http.StatusInternalServerError, "failed to delete "+rc.name)
			return
		}
		if res.RowsAffected == 0 {
			response.Error(c, http.StatusNotFound, rc.name+" not found")
			return
		}

		response.Success(c, gin.H{"deleted": id})
	}
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

func validateRequired(c *gin.Context, payload map[string]any, keys []string) bool {
	for _, k := range keys {
		v, ok := payload[k]
		if !ok || isEmpty(v) {
			response.Error(c, http.StatusBadRequest, k+" is required")
			return false
		}
	}
	return true
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
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

func programToResponse(p database.Program) gin.H {
	return gin.H{
		"id":         p.ID,
		"sasaran_id": p.SasaranID,
		"kode":       p.Kode,
		"nama":       p.Nama,
		"deskripsi":  p.Deskripsi,
		"created_at": p.CreatedAt,
		"updated_at": p.UpdatedAt,
	}
}

func mapWriteError(rc resourceConfig, action string, err error) string {
	msg := err.Error()
	if rc.path == "indikator_sub_kegiatan" {
		if strings.Contains(msg, "Unknown column") &&
			(strings.Contains(msg, "anggaran_tahun_sebelumnya") ||
				strings.Contains(msg, "anggaran_tahun_ini")) {
			return "schema indikator_sub_kegiatan belum mendukung field anggaran. Jalankan migration 015_add_anggaran_fields_to_indikator_sub_kegiatan.sql"
		}
	}

	return fmt.Sprintf("failed to %s %s: %v", action, rc.name, err)
}

func (h *Handler) defaultUnitPengusulID(c *gin.Context) (uint64, error) {
	var unit database.UnitPengusul
	err := h.db.WithContext(c.Request.Context()).Select("id").Order("id ASC").First(&unit).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.New("default unit_pengusul not found")
	}
	if err != nil {
		return 0, errors.New("failed to resolve default unit_pengusul")
	}
	return unit.ID, nil
}

func (h *Handler) defaultUnitPelaksanaID(c *gin.Context) (uint64, error) {
	var unit database.UnitPelaksana
	err := h.db.WithContext(c.Request.Context()).Select("id").Order("id ASC").First(&unit).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.New("default unit_pelaksana not found")
	}
	if err != nil {
		return 0, errors.New("failed to resolve default unit_pelaksana")
	}
	return unit.ID, nil
}

func planningResources() []resourceConfig {
	return []resourceConfig{
		{path: "misi", name: "misi", requiredKeys: []string{"visi_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.Misi{} }, newSlice: func() any { return &[]database.Misi{} }},
		{path: "tujuan", name: "tujuan", requiredKeys: []string{"misi_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.Tujuan{} }, newSlice: func() any { return &[]database.Tujuan{} }},
		{path: "indikator_tujuan", name: "indikator_tujuan", requiredKeys: []string{"tujuan_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorTujuan{} }, newSlice: func() any { return &[]database.IndikatorTujuan{} }},
		{path: "sasaran", name: "sasaran", requiredKeys: []string{"tujuan_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.Sasaran{} }, newSlice: func() any { return &[]database.Sasaran{} }},
		{path: "indikator_sasaran", name: "indikator_sasaran", requiredKeys: []string{"sasaran_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorSasaran{} }, newSlice: func() any { return &[]database.IndikatorSasaran{} }},
		{path: "program", name: "program", requiredKeys: []string{"sasaran_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.Program{} }, newSlice: func() any { return &[]database.Program{} }},
		{path: "indikator_program", name: "indikator_program", requiredKeys: []string{"program_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorProgram{} }, newSlice: func() any { return &[]database.IndikatorProgram{} }},
		{path: "kegiatan", name: "kegiatan", requiredKeys: []string{"program_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.Kegiatan{} }, newSlice: func() any { return &[]database.Kegiatan{} }},
		{path: "indikator_kegiatan", name: "indikator_kegiatan", requiredKeys: []string{"kegiatan_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorKegiatan{} }, newSlice: func() any { return &[]database.IndikatorKegiatan{} }},
		{path: "sub_kegiatan", name: "sub_kegiatan", requiredKeys: []string{"kegiatan_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.SubKegiatan{} }, newSlice: func() any { return &[]database.SubKegiatan{} }},
		{path: "indikator_sub_kegiatan", name: "indikator_sub_kegiatan", requiredKeys: []string{"sub_kegiatan_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorSubKegiatan{} }, newSlice: func() any { return &[]database.IndikatorSubKegiatan{} }},
		{path: "rencana_kerja", name: "rencana_kerja", requiredKeys: []string{"indikator_sub_kegiatan_id", "kode", "nama", "tahun", "unit_pengusul_id", "status", "dibuat_oleh"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.RencanaKerja{} }, newSlice: func() any { return &[]database.RencanaKerja{} }},
		{path: "indikator_rencana_kerja", name: "indikator_rencana_kerja", requiredKeys: []string{"rencana_kerja_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorRencanaKerja{} }, newSlice: func() any { return &[]database.IndikatorRencanaKerja{} }},
		{path: "realisasi_rencana_kerja", name: "realisasi_rencana_kerja", requiredKeys: []string{"indikator_rencana_kerja_id", "tahun", "nilai_realisasi", "realisasi_anggaran", "diinput_oleh"}, searchFields: []string{"keterangan"}, newModel: func() any { return &database.RealisasiRencanaKerja{} }, newSlice: func() any { return &[]database.RealisasiRencanaKerja{} }},
	}
}
