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
	"golang.org/x/crypto/bcrypt"
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

type userResponseItem struct {
	ID              uint64    `json:"id"`
	NamaLengkap     string    `json:"nama_lengkap"`
	Email           string    `json:"email"`
	Role            string    `json:"role"`
	Aktif           bool      `json:"aktif"`
	Status          string    `json:"status"`
	UnitPengusulID  *uint64   `json:"unit_pengusul_id"`
	UnitPengusulNama string   `json:"unit_pengusul_nama"`
	UnitPelaksanaID *uint64   `json:"unit_pelaksana_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type userUpsertPayload struct {
	NamaLengkap     string
	Email           string
	Role            string
	Aktif           bool
	UnitPengusulID  *uint64
	UnitPelaksanaID *uint64
	PasswordHash    string
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

		if rc.path == "users" {
			queryx := h.db.WithContext(c.Request.Context()).
				Table("users u").
				Select(strings.Join([]string{
					"u.id",
					"u.nama_lengkap",
					"u.email",
					"u.role",
					"u.aktif",
					"u.unit_pengusul_id",
					"u.unit_pelaksana_id",
					"COALESCE(up.nama, '') AS unit_pengusul_nama",
					"u.created_at",
					"u.updated_at",
				}, ", ")).
				Joins("LEFT JOIN unit_pengusul up ON up.id = u.unit_pengusul_id")

			if q != "" {
				like := "%" + q + "%"
				queryx = queryx.Where("u.nama_lengkap LIKE ? OR u.email LIKE ?", like, like)
			}

			if roleRaw := strings.TrimSpace(c.Query("role")); roleRaw != "" {
				role := strings.ToUpper(roleRaw)
				if !isValidUserRole(role) {
					response.Error(c, http.StatusBadRequest, "role harus salah satu: ADMIN, OPERATOR, PERENCANA, VERIFIKATOR, PIMPINAN")
					return
				}
				queryx = queryx.Where("u.role = ?", role)
			}

			if statusRaw := strings.TrimSpace(c.Query("status")); statusRaw != "" {
				aktif, err := parseStatusAktif(statusRaw)
				if err != nil {
					response.Error(c, http.StatusBadRequest, err.Error())
					return
				}
				queryx = queryx.Where("u.aktif = ?", aktif)
			}

			if unitRaw := strings.TrimSpace(c.Query("unit_pengusul_id")); unitRaw != "" {
				unitID, err := strconv.ParseUint(unitRaw, 10, 64)
				if err != nil || unitID == 0 {
					response.Error(c, http.StatusBadRequest, "unit_pengusul_id harus berupa angka > 0")
					return
				}
				queryx = queryx.Where("u.unit_pengusul_id = ?", unitID)
			}

			var rows []userResponseItem
			if err := queryx.Order("u.id DESC").Find(&rows).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
				return
			}

			for i := range rows {
				if rows[i].Aktif {
					rows[i].Status = "AKTIF"
				} else {
					rows[i].Status = "NONAKTIF"
				}
			}

			response.Success(c, gin.H{"items": rows, "total": len(rows)})
			return
		}

		if rc.path == "indikator_program" {
			var rows []database.IndikatorProgram
			qx := h.db.WithContext(c.Request.Context()).Model(&database.IndikatorProgram{}).Preload("Program")
			if q != "" {
				like := "%" + q + "%"
				qx = qx.Where("indikator_program.kode LIKE ? OR indikator_program.nama LIKE ?", like, like)
			}
			if err := qx.Order("indikator_program.id DESC").Find(&rows).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
				return
			}

			out := make([]gin.H, 0, len(rows))
			for _, ip := range rows {
				out = append(out, indikatorProgramToResponse(ip))
			}
			response.Success(c, gin.H{"items": out, "total": len(out)})
			return
		}
		if rc.path == "indikator_kegiatan" {
			sortBy := strings.TrimSpace(c.Query("sort_by"))
			order := strings.ToLower(strings.TrimSpace(c.Query("order")))
			if order != "asc" {
				order = "desc"
			}

			allowedSortColumns := map[string]string{
				"id":                   "indikator_kegiatan.id",
				"indikator_program_id": "indikator_kegiatan.indikator_program_id",
				"kegiatan_id":          "indikator_kegiatan.kegiatan_id",
				"kode":                 "indikator_kegiatan.kode",
				"nama":                 "indikator_kegiatan.nama",
				"baseline":             "indikator_kegiatan.baseline",
			}

			sortColumn, ok := allowedSortColumns[sortBy]
			if !ok {
				sortColumn = allowedSortColumns["id"]
			}

			queryx := h.db.WithContext(c.Request.Context()).
				Model(&database.IndikatorKegiatan{}).
				Joins("LEFT JOIN kegiatan ON kegiatan.id = indikator_kegiatan.kegiatan_id").
				Joins("LEFT JOIN indikator_program ON indikator_program.id = indikator_kegiatan.indikator_program_id").
				Preload("Kegiatan").
				Preload("IndikatorProgram")

			if q != "" {
				like := "%" + q + "%"
				queryx = queryx.Where(
					strings.Join([]string{
						"indikator_kegiatan.kode LIKE ?",
						"indikator_kegiatan.nama LIKE ?",
						"kegiatan.kode LIKE ?",
						"kegiatan.nama LIKE ?",
						"indikator_program.kode LIKE ?",
						"indikator_program.nama LIKE ?",
					}, " OR "),
					like, like, like, like, like, like,
				)
			}

			err := queryx.Order(sortColumn + " " + strings.ToUpper(order)).Find(items).Error
			if err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
				return
			}
			response.Success(c, gin.H{"items": items, "total": reflect.ValueOf(items).Elem().Len()})
			return
		}
		if rc.path == "indikator_sub_kegiatan" {
			sortBy := strings.TrimSpace(c.Query("sort_by"))
			order := strings.ToLower(strings.TrimSpace(c.Query("order")))
			if order != "asc" {
				order = "desc"
			}

			allowedSortColumns := map[string]string{
				"id":                   "indikator_sub_kegiatan.id",
				"indikator_kegiatan_id": "indikator_sub_kegiatan.indikator_kegiatan_id",
				"sub_kegiatan_id":      "indikator_sub_kegiatan.sub_kegiatan_id",
				"kode":                 "indikator_sub_kegiatan.kode",
				"nama":                 "indikator_sub_kegiatan.nama",
				"baseline":             "indikator_sub_kegiatan.baseline",
			}

			sortColumn, ok := allowedSortColumns[sortBy]
			if !ok {
				sortColumn = allowedSortColumns["id"]
			}

			queryx := h.db.WithContext(c.Request.Context()).
				Model(&database.IndikatorSubKegiatan{}).
				Joins("LEFT JOIN sub_kegiatan ON sub_kegiatan.id = indikator_sub_kegiatan.sub_kegiatan_id").
				Joins("LEFT JOIN indikator_kegiatan ON indikator_kegiatan.id = indikator_sub_kegiatan.indikator_kegiatan_id").
				Preload("SubKegiatan").
				Preload("IndikatorKegiatan")

			if q != "" {
				like := "%" + q + "%"
				queryx = queryx.Where(
					strings.Join([]string{
						"indikator_sub_kegiatan.kode LIKE ?",
						"indikator_sub_kegiatan.nama LIKE ?",
						"sub_kegiatan.kode LIKE ?",
						"sub_kegiatan.nama LIKE ?",
						"indikator_kegiatan.kode LIKE ?",
						"indikator_kegiatan.nama LIKE ?",
					}, " OR "),
					like, like, like, like, like, like,
				)
			}

			// ?all=true returns every record (for dropdowns / reference data)
			if c.Query("all") == "true" {
				err := queryx.Order(sortColumn + " " + strings.ToUpper(order)).Find(items).Error
				if err != nil {
					response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
					return
				}
				response.Success(c, gin.H{"items": items, "total": reflect.ValueOf(items).Elem().Len()})
				return
			}

			page := 1
			if v := strings.TrimSpace(c.Query("page")); v != "" {
				if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
					page = parsed
				}
			}

			limit := 5
			if v := strings.TrimSpace(c.Query("limit")); v != "" {
				if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
					limit = parsed
				}
			}
			if limit > 100 {
				limit = 100
			}

			var total int64
			if err := queryx.Session(&gorm.Session{}).Count(&total).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
				return
			}

			totalPages := int((total + int64(limit) - 1) / int64(limit))
			if totalPages < 1 {
				totalPages = 1
			}
			if page > totalPages {
				page = totalPages
			}

			offset := (page - 1) * limit
			err := queryx.Order(sortColumn + " " + strings.ToUpper(order)).Offset(offset).Limit(limit).Find(items).Error
			if err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
				return
			}
			response.Success(c, gin.H{
				"items":       items,
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
			})
			return
		}
		if rc.path == "rencana_kerja" {
			sortBy := strings.TrimSpace(c.Query("sort_by"))
			order := strings.ToLower(strings.TrimSpace(c.Query("order")))
			if order != "asc" {
				order = "desc"
			}

			allowedSortColumns := map[string]string{
				"id":                        "rencana_kerja.id",
				"indikator_sub_kegiatan_id": "rencana_kerja.indikator_sub_kegiatan_id",
				"unit_pengusul_id":          "rencana_kerja.unit_pengusul_id",
				"kode":                      "rencana_kerja.kode",
				"nama":                      "rencana_kerja.nama",
				"tahun":                     "rencana_kerja.tahun",
				"triwulan":                  "rencana_kerja.triwulan",
				"status":                    "rencana_kerja.status",
				"created_at":                "rencana_kerja.created_at",
				"updated_at":                "rencana_kerja.updated_at",
			}

			sortColumn, ok := allowedSortColumns[sortBy]
			if !ok {
				sortColumn = allowedSortColumns["id"]
			}

			queryx := h.db.WithContext(c.Request.Context()).
				Model(&database.RencanaKerja{}).
				Joins("LEFT JOIN indikator_sub_kegiatan ON indikator_sub_kegiatan.id = rencana_kerja.indikator_sub_kegiatan_id").
				Joins("LEFT JOIN unit_pengusul ON unit_pengusul.id = rencana_kerja.unit_pengusul_id").
				Preload("IndikatorSubKegiatan").
				Preload("UnitPengusul")

			if q != "" {
				like := "%" + q + "%"
				queryx = queryx.Where(
					strings.Join([]string{
						"rencana_kerja.kode LIKE ?",
						"rencana_kerja.nama LIKE ?",
						"indikator_sub_kegiatan.kode LIKE ?",
						"indikator_sub_kegiatan.nama LIKE ?",
						"unit_pengusul.kode LIKE ?",
						"unit_pengusul.nama LIKE ?",
					}, " OR "),
					like, like, like, like, like, like,
				)
			}

			if tahunRaw := strings.TrimSpace(c.Query("tahun")); tahunRaw != "" {
				tahun, err := strconv.Atoi(tahunRaw)
				if err != nil || tahun < 2000 || tahun > 2100 {
					response.Error(c, http.StatusBadRequest, "tahun harus berupa angka antara 2000-2100")
					return
				}
				queryx = queryx.Where("rencana_kerja.tahun = ?", tahun)
			}

			if triwulanRaw := strings.TrimSpace(c.Query("triwulan")); triwulanRaw != "" {
				triwulan, err := strconv.Atoi(triwulanRaw)
				if err != nil || triwulan < 1 || triwulan > 4 {
					response.Error(c, http.StatusBadRequest, "triwulan harus berupa angka antara 1-4")
					return
				}
				queryx = queryx.Where("rencana_kerja.triwulan = ?", triwulan)
			}

			if unitIDRaw := strings.TrimSpace(c.Query("unit_pengusul_id")); unitIDRaw != "" {
				unitID, err := strconv.ParseUint(unitIDRaw, 10, 64)
				if err != nil || unitID == 0 {
					response.Error(c, http.StatusBadRequest, "unit_pengusul_id harus berupa angka > 0")
					return
				}
				queryx = queryx.Where("rencana_kerja.unit_pengusul_id = ?", unitID)
			}

			if indikatorIDRaw := strings.TrimSpace(c.Query("indikator_sub_kegiatan_id")); indikatorIDRaw != "" {
				indikatorID, err := strconv.ParseUint(indikatorIDRaw, 10, 64)
				if err != nil || indikatorID == 0 {
					response.Error(c, http.StatusBadRequest, "indikator_sub_kegiatan_id harus berupa angka > 0")
					return
				}
				queryx = queryx.Where("rencana_kerja.indikator_sub_kegiatan_id = ?", indikatorID)
			}

			if statusRaw := strings.TrimSpace(c.Query("status")); statusRaw != "" {
				status := strings.ToUpper(statusRaw)
				if !isValidRencanaKerjaStatus(status) {
					response.Error(c, http.StatusBadRequest, "status harus salah satu: DRAFT, DIAJUKAN, DISETUJUI, DITOLAK")
					return
				}
				queryx = queryx.Where("rencana_kerja.status = ?", status)
			}

			if isTruthy(c.Query("final_only")) {
				queryx = queryx.Where("rencana_kerja.status = ?", "DISETUJUI")
			}

			if c.Query("all") == "true" {
				err := queryx.Order(sortColumn + " " + strings.ToUpper(order)).Find(items).Error
				if err != nil {
					response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
					return
				}
				response.Success(c, gin.H{"items": items, "total": reflect.ValueOf(items).Elem().Len()})
				return
			}

			page := 1
			if v := strings.TrimSpace(c.Query("page")); v != "" {
				if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
					page = parsed
				}
			}

			limit := 10
			if v := strings.TrimSpace(c.Query("limit")); v != "" {
				if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
					limit = parsed
				}
			}
			if limit > 100 {
				limit = 100
			}

			var total int64
			if err := queryx.Session(&gorm.Session{}).Count(&total).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
				return
			}

			totalPages := int((total + int64(limit) - 1) / int64(limit))
			if totalPages < 1 {
				totalPages = 1
			}
			if page > totalPages {
				page = totalPages
			}

			offset := (page - 1) * limit
			err := queryx.Order(sortColumn + " " + strings.ToUpper(order)).Offset(offset).Limit(limit).Find(items).Error
			if err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
				return
			}

			response.Success(c, gin.H{
				"items":       items,
				"total":       total,
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
			})
			return
		}
		if rc.path == "indikator_rencana_kerja" && isTruthy(c.Query("final_only")) {
			query = query.Joins("JOIN rencana_kerja rk ON rk.id = indikator_rencana_kerja.rencana_kerja_id").
				Where("rk.status = ?", "DISETUJUI")
		}
		if rc.path == "realisasi_rencana_kerja" && isTruthy(c.Query("final_only")) {
			query = query.Joins("JOIN indikator_rencana_kerja irk ON irk.id = realisasi_rencana_kerja.indikator_rencana_kerja_id").
				Joins("JOIN rencana_kerja rk ON rk.id = irk.rencana_kerja_id").
				Where("rk.status = ?", "DISETUJUI")
		}
		if rc.path == "pagu_sub_kegiatan" {
			tahunRaw := strings.TrimSpace(c.Query("tahun"))
			if tahunRaw != "" {
				tahun, err := strconv.Atoi(tahunRaw)
				if err != nil || tahun < 2000 || tahun > 2100 {
					response.Error(c, http.StatusBadRequest, "tahun harus berupa angka antara 2000-2100")
					return
				}
				query = query.Where("tahun = ?", tahun)
			}
			query = query.Preload("SubKegiatan")
		}
		if err := query.Order("id DESC").Find(items).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, mapReadError(rc, "list", err))
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

		if rc.path == "users" {
			var row userResponseItem
			err := h.db.WithContext(c.Request.Context()).
				Table("users u").
				Select(strings.Join([]string{
					"u.id",
					"u.nama_lengkap",
					"u.email",
					"u.role",
					"u.aktif",
					"u.unit_pengusul_id",
					"u.unit_pelaksana_id",
					"COALESCE(up.nama, '') AS unit_pengusul_nama",
					"u.created_at",
					"u.updated_at",
				}, ", ")).
				Joins("LEFT JOIN unit_pengusul up ON up.id = u.unit_pengusul_id").
				Where("u.id = ?", id).
				First(&row).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusNotFound, rc.name+" not found")
				return
			}
			if err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}

			if row.Aktif {
				row.Status = "AKTIF"
			} else {
				row.Status = "NONAKTIF"
			}

			response.Success(c, row)
			return
		}

		if rc.path == "indikator_program" {
			var item database.IndikatorProgram
			err := h.db.WithContext(c.Request.Context()).
				Preload("Program").
				First(&item, "id = ?", id).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusNotFound, rc.name+" not found")
				return
			}
			if err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}

			response.Success(c, indikatorProgramToResponse(item))
			return
		}
		if rc.path == "indikator_sub_kegiatan" {
			var item database.IndikatorSubKegiatan
			err := h.db.WithContext(c.Request.Context()).
				Preload("SubKegiatan").
				Preload("IndikatorKegiatan").
				First(&item, "id = ?", id).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusNotFound, rc.name+" not found")
				return
			}
			if err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}

			response.Success(c, item)
			return
		}
		if rc.path == "rencana_kerja" {
			var item database.RencanaKerja
			err := h.db.WithContext(c.Request.Context()).
				Preload("IndikatorSubKegiatan").
				Preload("UnitPengusul").
				First(&item, "id = ?", id).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusNotFound, rc.name+" not found")
				return
			}
			if err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}

			response.Success(c, item)
			return
		}
		if rc.path == "pagu_sub_kegiatan" {
			var item database.PaguSubKegiatan
			err := h.db.WithContext(c.Request.Context()).
				Preload("SubKegiatan").
				First(&item, "id = ?", id).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusNotFound, rc.name+" not found")
				return
			}
			if err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}

			response.Success(c, item)
			return
		}

		item := rc.newModel()
		err := h.db.WithContext(c.Request.Context()).First(item, "id = ?", id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, rc.name+" not found")
			return
		}
		if err != nil {
			response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
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

		if !allowedToMutatePlanningData(c, rc.path) {
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

		if rc.path == "users" {
			userPayload, err := userFromPayload(payload, true)
			if err != nil {
				response.Error(c, http.StatusBadRequest, err.Error())
				return
			}

			user := database.User{
				UnitPengusulID:  userPayload.UnitPengusulID,
				UnitPelaksanaID: userPayload.UnitPelaksanaID,
				NamaLengkap:     userPayload.NamaLengkap,
				Email:           userPayload.Email,
				PasswordHash:    userPayload.PasswordHash,
				Role:            userPayload.Role,
				Aktif:           userPayload.Aktif,
			}

			now := time.Now()
			user.CreatedAt = now
			user.UpdatedAt = now

			if err := h.db.WithContext(c.Request.Context()).Create(&user).Error; err != nil {
				response.Error(c, http.StatusBadRequest, mapWriteError(rc, "create", err))
				return
			}

			var created userResponseItem
			if err := h.db.WithContext(c.Request.Context()).
				Table("users u").
				Select(strings.Join([]string{
					"u.id",
					"u.nama_lengkap",
					"u.email",
					"u.role",
					"u.aktif",
					"u.unit_pengusul_id",
					"u.unit_pelaksana_id",
					"COALESCE(up.nama, '') AS unit_pengusul_nama",
					"u.created_at",
					"u.updated_at",
				}, ", ")).
				Joins("LEFT JOIN unit_pengusul up ON up.id = u.unit_pengusul_id").
				Where("u.id = ?", user.ID).
				First(&created).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}
			if created.Aktif {
				created.Status = "AKTIF"
			} else {
				created.Status = "NONAKTIF"
			}

			response.Success(c, created)
			return
		}

		if rc.path == "realisasi_rencana_kerja" {
			indikatorID, err := payloadUint64(payload, "indikator_rencana_kerja_id")
			if err != nil || indikatorID == 0 {
				response.Error(c, http.StatusBadRequest, "indikator_rencana_kerja_id harus berupa angka > 0")
				return
			}
			if err := h.ensureIndikatorRencanaKerjaFinal(c, indikatorID); err != nil {
				response.Error(c, http.StatusUnprocessableEntity, err.Error())
				return
			}
		}

		if rc.path == "rencana_kerja" {
			if err := normalizeRencanaKerjaPayload(payload); err != nil {
				response.Error(c, http.StatusBadRequest, err.Error())
				return
			}

			rencanaKerja, err := rencanaKerjaFromPayload(payload)
			if err != nil {
				response.Error(c, http.StatusBadRequest, err.Error())
				return
			}

			now := time.Now()
			rencanaKerja.CreatedAt = now
			rencanaKerja.UpdatedAt = now

			if err := h.db.WithContext(c.Request.Context()).Create(&rencanaKerja).Error; err != nil {
				response.Error(c, http.StatusBadRequest, mapWriteError(rc, "create", err))
				return
			}

			var created database.RencanaKerja
			if err := h.db.WithContext(c.Request.Context()).
				Preload("IndikatorSubKegiatan").
				Preload("UnitPengusul").
				First(&created, "id = ?", rencanaKerja.ID).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}

			response.Success(c, created)
			return
		}

		if rc.path == "program" {
			// Program API hanya menerima kode, nama, deskripsi, created_at, updated_at
			delete(payload, "unit_pengusul_id")
			delete(payload, "sasaran_id")
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

		if supportsTimestamps(rc.newModel()) {
			now := time.Now()
			if _, ok := payload["created_at"]; !ok {
				payload["created_at"] = now
			}
			if _, ok := payload["updated_at"]; !ok {
				payload["updated_at"] = now
			}
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

		if !allowedToMutatePlanningData(c, rc.path) {
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
		if rc.path == "users" {
			userPayload, err := userFromPayload(payload, false)
			if err != nil {
				response.Error(c, http.StatusBadRequest, err.Error())
				return
			}

			updates := map[string]any{}
			if userPayload.NamaLengkap != "" {
				updates["nama_lengkap"] = userPayload.NamaLengkap
			}
			if userPayload.Email != "" {
				updates["email"] = userPayload.Email
			}
			if userPayload.Role != "" {
				updates["role"] = userPayload.Role
			}
			updates["aktif"] = userPayload.Aktif
			updates["unit_pengusul_id"] = userPayload.UnitPengusulID
			updates["unit_pelaksana_id"] = userPayload.UnitPelaksanaID
			if userPayload.PasswordHash != "" {
				updates["password_hash"] = userPayload.PasswordHash
			}

			if len(updates) == 0 {
				response.Error(c, http.StatusBadRequest, "tidak ada field yang dapat diperbarui")
				return
			}

			updates["updated_at"] = time.Now()

			res := h.db.WithContext(c.Request.Context()).Model(&database.User{}).Where("id = ?", id).Updates(updates)
			if res.Error != nil {
				response.Error(c, http.StatusBadRequest, mapWriteError(rc, "update", res.Error))
				return
			}
			if res.RowsAffected == 0 {
				response.Error(c, http.StatusNotFound, rc.name+" not found")
				return
			}

			var updated userResponseItem
			if err := h.db.WithContext(c.Request.Context()).
				Table("users u").
				Select(strings.Join([]string{
					"u.id",
					"u.nama_lengkap",
					"u.email",
					"u.role",
					"u.aktif",
					"u.unit_pengusul_id",
					"u.unit_pelaksana_id",
					"COALESCE(up.nama, '') AS unit_pengusul_nama",
					"u.created_at",
					"u.updated_at",
				}, ", ")).
				Joins("LEFT JOIN unit_pengusul up ON up.id = u.unit_pengusul_id").
				Where("u.id = ?", id).
				First(&updated).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}
			if updated.Aktif {
				updated.Status = "AKTIF"
			} else {
				updated.Status = "NONAKTIF"
			}

			response.Success(c, updated)
			return
		}
		if rc.path == "realisasi_rencana_kerja" {
			indikatorID, err := payloadUint64(payload, "indikator_rencana_kerja_id")
			if err != nil || indikatorID == 0 {
				response.Error(c, http.StatusBadRequest, "indikator_rencana_kerja_id harus berupa angka > 0")
				return
			}
			if err := h.ensureIndikatorRencanaKerjaFinal(c, indikatorID); err != nil {
				response.Error(c, http.StatusUnprocessableEntity, err.Error())
				return
			}
		}
		if rc.path == "rencana_kerja" {
			if err := normalizeRencanaKerjaPayload(payload); err != nil {
				response.Error(c, http.StatusBadRequest, err.Error())
				return
			}

			delete(payload, "id")
			delete(payload, "created_at")

			res := h.db.WithContext(c.Request.Context()).
				Model(&database.RencanaKerja{}).
				Where("id = ?", id).
				Updates(payload)
			if res.Error != nil {
				response.Error(c, http.StatusBadRequest, mapWriteError(rc, "update", res.Error))
				return
			}
			if res.RowsAffected == 0 {
				response.Error(c, http.StatusNotFound, rc.name+" not found")
				return
			}

			var updated database.RencanaKerja
			if err := h.db.WithContext(c.Request.Context()).
				Preload("IndikatorSubKegiatan").
				Preload("UnitPengusul").
				First(&updated, "id = ?", id).Error; err != nil {
				response.Error(c, http.StatusInternalServerError, mapReadError(rc, "get", err))
				return
			}

			response.Success(c, updated)
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

		if !allowedToMutatePlanningData(c, rc.path) {
			return
		}

		id, ok := parseID(c)
		if !ok {
			return
		}

		res := h.db.WithContext(c.Request.Context()).Delete(rc.newModel(), "id = ?", id)
		if res.Error != nil {
			if rc.path == "rencana_kerja" {
				response.Error(c, http.StatusBadRequest, mapWriteError(rc, "delete", res.Error))
				return
			}
			response.Error(c, http.StatusInternalServerError, mapReadError(rc, "delete", res.Error))
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

// planningMutationAllowedRoles are the roles permitted to mutate any planning resource.
var planningMutationAllowedRoles = map[string]bool{
	"ADMIN":     true,
	"OPERATOR":  true,
	"PERENCANA": true,
}

// allowedToMutatePlanningData returns false (with 403) when the caller's role
// is not in planningMutationAllowedRoles. For the "users" resource, only ADMIN
// is allowed.
func allowedToMutatePlanningData(c *gin.Context, resourcePath string) bool {
	rawRole, _ := c.Get("auth.role")
	role := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", rawRole)))
	if resourcePath == "users" {
		if role != "ADMIN" {
			response.Error(c, http.StatusForbidden, "hanya ADMIN yang dapat mengelola data pengguna")
			return false
		}
		return true
	}
	if !planningMutationAllowedRoles[role] {
		response.Error(c, http.StatusForbidden, "anda tidak memiliki hak akses untuk melakukan perubahan data perencanaan")
		return false
	}
	return true
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

func supportsTimestamps(model any) bool {
	t := reflect.TypeOf(model)
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	_, hasCreatedAt := t.FieldByName("CreatedAt")
	_, hasUpdatedAt := t.FieldByName("UpdatedAt")
	return hasCreatedAt && hasUpdatedAt
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
		"kode":       p.Kode,
		"nama":       p.Nama,
		"deskripsi":  p.Deskripsi,
		"created_at": p.CreatedAt,
		"updated_at": p.UpdatedAt,
	}
}

func indikatorProgramToResponse(ip database.IndikatorProgram) gin.H {
	programNama := ""
	programKode := ""
	if ip.Program != nil {
		programNama = ip.Program.Nama
		programKode = ip.Program.Kode
	}
	return gin.H{
		"id":          ip.ID,
		"sasaran_id":  ip.SasaranID,
		"program_id":  ip.ProgramID,
		"program_nama": programNama,
		"program_kode": programKode,
		"kode":        ip.Kode,
		"nama":        ip.Nama,
		"formula":     ip.Formula,
		"satuan":      ip.Satuan,
		"baseline":    ip.Baseline,
		"created_at":  ip.CreatedAt,
		"updated_at":  ip.UpdatedAt,
	}
}

func mapWriteError(rc resourceConfig, action string, err error) string {
	msg := err.Error()
	if rc.path == "users" {
		if strings.Contains(msg, "Duplicate entry") && strings.Contains(msg, "uq_users_email") {
			return "email user sudah digunakan"
		}
		if strings.Contains(msg, "Data truncated") && strings.Contains(msg, "role") {
			return "role user tidak valid"
		}
		if strings.Contains(msg, "Unknown column") && strings.Contains(msg, "nama_lengkap") {
			return "schema users tidak sesuai. Pastikan menggunakan schema performance (nama_lengkap, aktif, role)"
		}
	}
	if rc.path == "rencana_kerja" {
		if strings.Contains(msg, "Duplicate entry") && strings.Contains(msg, "uq_rencana_kerja_kode") {
			return "kode rencana_kerja sudah digunakan"
		}
		if action == "delete" && strings.Contains(msg, "foreign key constraint fails") && strings.Contains(msg, "indikator_rencana_kerja") {
			return "rencana_kerja tidak dapat dihapus karena masih dipakai oleh indikator_rencana_kerja"
		}
		if strings.Contains(msg, "foreign key constraint fails") && strings.Contains(msg, "indikator_sub_kegiatan") {
			return "indikator_sub_kegiatan_id tidak valid"
		}
		if strings.Contains(msg, "foreign key constraint fails") && strings.Contains(msg, "unit_pengusul") {
			return "unit_pengusul_id tidak valid"
		}
	}
	if rc.path == "indikator_kegiatan" {
		if strings.Contains(msg, "Duplicate entry") && strings.Contains(msg, "uq_indikator_kegiatan_kode") {
			return "kode indikator_kegiatan sudah digunakan"
		}
	}
	if rc.path == "indikator_rencana_kerja" {
		if strings.Contains(msg, "Duplicate entry") && strings.Contains(msg, "uq_indikator_rk_kode") {
			return "kode indikator_rencana_kerja sudah digunakan"
		}
	}
	if rc.path == "indikator_sub_kegiatan" {
		if strings.Contains(msg, "Duplicate entry") && strings.Contains(msg, "uq_indikator_sub_kegiatan_kode") {
			return "kode indikator_sub_kegiatan sudah digunakan"
		}
		if strings.Contains(msg, "Unknown column") && strings.Contains(msg, "indikator_kegiatan_id") {
			return "schema indikator_sub_kegiatan belum mendukung relasi indikator_kegiatan_id. Jalankan migration 023 dan 024"
		}
		if strings.Contains(msg, "Unknown column") &&
			(strings.Contains(msg, "anggaran_tahun_sebelumnya") ||
				strings.Contains(msg, "anggaran_tahun_ini")) {
			return "schema indikator_sub_kegiatan belum mendukung field anggaran. Jalankan migration 015_add_anggaran_fields_to_indikator_sub_kegiatan.sql"
		}
	}

	return fmt.Sprintf("failed to %s %s: %v", action, rc.name, err)
}

func mapReadError(rc resourceConfig, action string, err error) string {
	msg := err.Error()
	if rc.path == "pagu_sub_kegiatan" {
		if strings.Contains(msg, "doesn't exist") && strings.Contains(msg, "pagu_sub_kegiatan") {
			return "schema pagu_sub_kegiatan belum tersedia. Jalankan migration 017_create_pagu_sub_kegiatan.sql"
		}
	}

	return fmt.Sprintf("failed to %s %s: %v", action, rc.name, err)
}

func normalizeRencanaKerjaPayload(payload map[string]any) error {
	if tahunRaw, ok := payload["tahun"]; ok {
		tahun, err := coerceInt(tahunRaw)
		if err != nil || tahun < 2000 || tahun > 2100 {
			return errors.New("tahun harus berupa angka antara 2000-2100")
		}
		payload["tahun"] = tahun
	}

	if triwulanRaw, ok := payload["triwulan"]; ok {
		if triwulanRaw == nil || strings.TrimSpace(fmt.Sprintf("%v", triwulanRaw)) == "" {
			delete(payload, "triwulan")
		} else {
			triwulan, err := coerceInt(triwulanRaw)
			if err != nil || triwulan < 1 || triwulan > 4 {
				return errors.New("triwulan harus berupa angka antara 1-4")
			}
			payload["triwulan"] = triwulan
		}
	}

	if statusRaw, ok := payload["status"]; ok {
		status := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", statusRaw)))
		if !isValidRencanaKerjaStatus(status) {
			return errors.New("status harus salah satu: DRAFT, DIAJUKAN, DISETUJUI, DITOLAK")
		}
		payload["status"] = status
	}

	for _, key := range []string{"indikator_sub_kegiatan_id", "unit_pengusul_id", "dibuat_oleh"} {
		if raw, ok := payload[key]; ok {
			v, err := coerceInt(raw)
			if err != nil || v <= 0 {
				return fmt.Errorf("%s harus berupa angka > 0", key)
			}
			payload[key] = v
		}
	}

	return nil
}

func isValidRencanaKerjaStatus(status string) bool {
	switch status {
	case "DRAFT", "DIAJUKAN", "DISETUJUI", "DITOLAK":
		return true
	default:
		return false
	}
}

func coerceInt(v any) (int, error) {
	switch n := v.(type) {
	case int:
		return n, nil
	case int8:
		return int(n), nil
	case int16:
		return int(n), nil
	case int32:
		return int(n), nil
	case int64:
		return int(n), nil
	case uint:
		return int(n), nil
	case uint8:
		return int(n), nil
	case uint16:
		return int(n), nil
	case uint32:
		return int(n), nil
	case uint64:
		return int(n), nil
	case float32:
		return int(n), nil
	case float64:
		return int(n), nil
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return 0, errors.New("empty value")
		}
		parsed, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

func isTruthy(v string) bool {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func payloadUint64(payload map[string]any, key string) (uint64, error) {
	raw, ok := payload[key]
	if !ok {
		return 0, fmt.Errorf("%s is required", key)
	}
	v, err := coerceInt(raw)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return uint64(v), nil
}

func isValidUserRole(role string) bool {
	switch role {
	case "ADMIN", "OPERATOR", "PERENCANA", "VERIFIKATOR", "PIMPINAN":
		return true
	default:
		return false
	}
}

func parseStatusAktif(statusRaw string) (bool, error) {
	status := strings.ToUpper(strings.TrimSpace(statusRaw))
	switch status {
	case "AKTIF":
		return true, nil
	case "NONAKTIF":
		return false, nil
	default:
		return false, errors.New("status harus AKTIF atau NONAKTIF")
	}
}

func boolFromAny(v any) (bool, error) {
	switch b := v.(type) {
	case bool:
		return b, nil
	case string:
		s := strings.ToLower(strings.TrimSpace(b))
		switch s {
		case "1", "true", "yes", "y", "on", "aktif":
			return true, nil
		case "0", "false", "no", "n", "off", "nonaktif":
			return false, nil
		default:
			return false, errors.New("invalid bool string")
		}
	case int, int8, int16, int32, int64:
		iv, err := coerceInt(v)
		if err != nil {
			return false, err
		}
		return iv != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		iv, err := coerceInt(v)
		if err != nil {
			return false, err
		}
		return iv != 0, nil
	case float32, float64:
		iv, err := coerceInt(v)
		if err != nil {
			return false, err
		}
		return iv != 0, nil
	default:
		return false, fmt.Errorf("unsupported type %T", v)
	}
}

func userFromPayload(payload map[string]any, isCreate bool) (userUpsertPayload, error) {
	out := userUpsertPayload{Aktif: true}

	if raw, ok := payload["nama_lengkap"]; ok {
		out.NamaLengkap = strings.TrimSpace(fmt.Sprintf("%v", raw))
	}
	if isCreate && out.NamaLengkap == "" {
		return userUpsertPayload{}, errors.New("nama_lengkap wajib diisi")
	}

	if raw, ok := payload["email"]; ok {
		out.Email = strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", raw)))
	}
	if isCreate && out.Email == "" {
		return userUpsertPayload{}, errors.New("email wajib diisi")
	}

	if raw, ok := payload["role"]; ok {
		out.Role = strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", raw)))
	}
	if isCreate && out.Role == "" {
		out.Role = "PERENCANA"
	}
	if out.Role != "" && !isValidUserRole(out.Role) {
		return userUpsertPayload{}, errors.New("role harus salah satu: ADMIN, OPERATOR, PERENCANA, VERIFIKATOR, PIMPINAN")
	}

	if raw, ok := payload["status"]; ok {
		aktif, err := parseStatusAktif(fmt.Sprintf("%v", raw))
		if err != nil {
			return userUpsertPayload{}, err
		}
		out.Aktif = aktif
	} else if raw, ok := payload["aktif"]; ok {
		aktif, err := boolFromAny(raw)
		if err != nil {
			return userUpsertPayload{}, errors.New("aktif harus berupa boolean")
		}
		out.Aktif = aktif
	}

	if raw, ok := payload["unit_pengusul_id"]; ok {
		if raw == nil || strings.TrimSpace(fmt.Sprintf("%v", raw)) == "" {
			out.UnitPengusulID = nil
		} else {
			v, err := coerceInt(raw)
			if err != nil || v <= 0 {
				return userUpsertPayload{}, errors.New("unit_pengusul_id harus berupa angka > 0")
			}
			uid := uint64(v)
			out.UnitPengusulID = &uid
		}
	}

	if raw, ok := payload["unit_pelaksana_id"]; ok {
		if raw == nil || strings.TrimSpace(fmt.Sprintf("%v", raw)) == "" {
			out.UnitPelaksanaID = nil
		} else {
			v, err := coerceInt(raw)
			if err != nil || v <= 0 {
				return userUpsertPayload{}, errors.New("unit_pelaksana_id harus berupa angka > 0")
			}
			uid := uint64(v)
			out.UnitPelaksanaID = &uid
		}
	}

	if raw, ok := payload["password"]; ok {
		password := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if password != "" {
			if len(password) < 8 {
				return userUpsertPayload{}, errors.New("password minimal 8 karakter")
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return userUpsertPayload{}, errors.New("gagal memproses password")
			}
			out.PasswordHash = string(hash)
		}
	}

	if isCreate && out.PasswordHash == "" {
		return userUpsertPayload{}, errors.New("password wajib diisi")
	}

	return out, nil
}

func (h *Handler) ensureIndikatorRencanaKerjaFinal(c *gin.Context, indikatorRencanaKerjaID uint64) error {
	var total int64
	err := h.db.WithContext(c.Request.Context()).
		Table("indikator_rencana_kerja irk").
		Joins("JOIN rencana_kerja rk ON rk.id = irk.rencana_kerja_id").
		Where("irk.id = ? AND rk.status = ?", indikatorRencanaKerjaID, "DISETUJUI").
		Count(&total).Error
	if err != nil {
		return errors.New("gagal memverifikasi status rencana_kerja")
	}
	if total == 0 {
		return errors.New("Evaluasi Rencana Aksi hanya dapat digunakan untuk rencana_kerja berstatus DISETUJUI")
	}
	return nil
}

func rencanaKerjaFromPayload(payload map[string]any) (database.RencanaKerja, error) {
	indikatorSubKegiatanID, err := coerceInt(payload["indikator_sub_kegiatan_id"])
	if err != nil || indikatorSubKegiatanID <= 0 {
		return database.RencanaKerja{}, errors.New("indikator_sub_kegiatan_id harus berupa angka > 0")
	}

	unitPengusulID, err := coerceInt(payload["unit_pengusul_id"])
	if err != nil || unitPengusulID <= 0 {
		return database.RencanaKerja{}, errors.New("unit_pengusul_id harus berupa angka > 0")
	}

	tahun, err := coerceInt(payload["tahun"])
	if err != nil || tahun < 2000 || tahun > 2100 {
		return database.RencanaKerja{}, errors.New("tahun harus berupa angka antara 2000-2100")
	}

	dibuatOleh, err := coerceInt(payload["dibuat_oleh"])
	if err != nil || dibuatOleh <= 0 {
		return database.RencanaKerja{}, errors.New("dibuat_oleh harus berupa angka > 0")
	}

	status := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", payload["status"])))
	if !isValidRencanaKerjaStatus(status) {
		return database.RencanaKerja{}, errors.New("status harus salah satu: DRAFT, DIAJUKAN, DISETUJUI, DITOLAK")
	}

	kode := strings.TrimSpace(fmt.Sprintf("%v", payload["kode"]))
	nama := strings.TrimSpace(fmt.Sprintf("%v", payload["nama"]))
	if kode == "" || nama == "" {
		return database.RencanaKerja{}, errors.New("kode dan nama wajib diisi")
	}

	model := database.RencanaKerja{
		IndikatorSubKegiatanID: uint64(indikatorSubKegiatanID),
		UnitPengusulID:         uint64(unitPengusulID),
		Kode:                   kode,
		Nama:                   nama,
		Tahun:                  int16(tahun),
		Status:                 status,
		DibuatOleh:             uint64(dibuatOleh),
	}

	if triwulanRaw, ok := payload["triwulan"]; ok {
		if triwulanRaw != nil && strings.TrimSpace(fmt.Sprintf("%v", triwulanRaw)) != "" {
			triwulan, err := coerceInt(triwulanRaw)
			if err != nil || triwulan < 1 || triwulan > 4 {
				return database.RencanaKerja{}, errors.New("triwulan harus berupa angka antara 1-4")
			}
			triwulanInt8 := int8(triwulan)
			model.Triwulan = &triwulanInt8
		}
	}

	if catatanRaw, ok := payload["catatan"]; ok {
		model.Catatan = strings.TrimSpace(fmt.Sprintf("%v", catatanRaw))
	}

	return model, nil
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
		{path: "program", name: "program", requiredKeys: []string{"kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.Program{} }, newSlice: func() any { return &[]database.Program{} }},
		{path: "indikator_program", name: "indikator_program", requiredKeys: []string{"sasaran_id", "program_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorProgram{} }, newSlice: func() any { return &[]database.IndikatorProgram{} }},
		{path: "kegiatan", name: "kegiatan", requiredKeys: []string{"program_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.Kegiatan{} }, newSlice: func() any { return &[]database.Kegiatan{} }},
		{path: "indikator_kegiatan", name: "indikator_kegiatan", requiredKeys: []string{"indikator_program_id", "kegiatan_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorKegiatan{} }, newSlice: func() any { return &[]database.IndikatorKegiatan{} }},
		{path: "sub_kegiatan", name: "sub_kegiatan", requiredKeys: []string{"kegiatan_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.SubKegiatan{} }, newSlice: func() any { return &[]database.SubKegiatan{} }},
		{path: "pagu_sub_kegiatan", name: "pagu_sub_kegiatan", requiredKeys: []string{"sub_kegiatan_id", "tahun", "pagu_tahun_sebelumnya", "pagu_tahun_ini"}, searchFields: []string{"CAST(sub_kegiatan_id AS CHAR)", "CAST(tahun AS CHAR)"}, newModel: func() any { return &database.PaguSubKegiatan{} }, newSlice: func() any { return &[]database.PaguSubKegiatan{} }},
		{path: "indikator_sub_kegiatan", name: "indikator_sub_kegiatan", requiredKeys: []string{"indikator_kegiatan_id", "sub_kegiatan_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorSubKegiatan{} }, newSlice: func() any { return &[]database.IndikatorSubKegiatan{} }},
		{path: "rencana_kerja", name: "rencana_kerja", requiredKeys: []string{"indikator_sub_kegiatan_id", "kode", "nama", "tahun", "unit_pengusul_id", "status", "dibuat_oleh"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.RencanaKerja{} }, newSlice: func() any { return &[]database.RencanaKerja{} }},
		{path: "indikator_rencana_kerja", name: "indikator_rencana_kerja", requiredKeys: []string{"rencana_kerja_id", "kode", "nama"}, searchFields: []string{"kode", "nama"}, newModel: func() any { return &database.IndikatorRencanaKerja{} }, newSlice: func() any { return &[]database.IndikatorRencanaKerja{} }},
		{path: "realisasi_rencana_kerja", name: "realisasi_rencana_kerja", requiredKeys: []string{"indikator_rencana_kerja_id", "tahun", "nilai_realisasi", "realisasi_anggaran", "diinput_oleh"}, searchFields: []string{"keterangan"}, newModel: func() any { return &database.RealisasiRencanaKerja{} }, newSlice: func() any { return &[]database.RealisasiRencanaKerja{} }},
		{path: "users", name: "users", requiredKeys: []string{"nama_lengkap", "email", "role"}, searchFields: []string{"nama_lengkap", "email"}, newModel: func() any { return &database.User{} }, newSlice: func() any { return &[]database.User{} }},
	}
}
