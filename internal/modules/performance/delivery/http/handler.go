package http

import (
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type updateTargetRealisasiRequest struct {
	TargetNilai       *float64 `json:"target_nilai"`
	RealisasiNilai    *float64 `json:"realisasi_nilai"`
	Status            string   `json:"status"`
	Catatan           string   `json:"catatan"`
	DiverifikasiOleh  *uint64  `json:"diverifikasi_oleh"`
	TanggalVerifikasi string   `json:"tanggal_verifikasi"`
}

type createTargetRealisasiRequest struct {
	IndikatorRencanaKerjaID uint64  `json:"indikator_rencana_kerja_id"`
	Tahun                   int16   `json:"tahun"`
	Triwulan                int8    `json:"triwulan"`
	TargetNilai             float64 `json:"target_nilai"`
	RealisasiNilai          float64 `json:"realisasi_nilai"`
	Status                  string  `json:"status"`
	Catatan                 string  `json:"catatan"`
	DiverifikasiOleh        *uint64 `json:"diverifikasi_oleh"`
	TanggalVerifikasi       string  `json:"tanggal_verifikasi"`
}

type calculateAchievementRequest struct {
	TargetNilai    float64 `json:"target_nilai"`
	RealisasiNilai float64 `json:"realisasi_nilai"`
}

type informasiUpsertRequest struct {
	Informasi                 string `json:"informasi" form:"informasi" binding:"required"`
	Tahun                     int    `json:"tahun" form:"tahun" binding:"required"`
	PilihanRouteHalamanTujuan string `json:"pilihan_route_halaman_tujuan" form:"pilihan_route_halaman_tujuan" binding:"required"`
}

type chartTargetVsRealisasiRow struct {
	Tahun          int     `gorm:"column:tahun"`
	Triwulan       int     `gorm:"column:triwulan"`
	TotalTarget    float64 `gorm:"column:total_target"`
	TotalRealisasi float64 `gorm:"column:total_realisasi"`
}

type yearlyPerformanceSummaryRow struct {
	Tahun               int     `gorm:"column:tahun"`
	TotalData           int64   `gorm:"column:total_data"`
	TotalIndikator      int64   `gorm:"column:total_indikator"`
	TotalTargetNilai    float64 `gorm:"column:total_target_nilai"`
	TotalRealisasiNilai float64 `gorm:"column:total_realisasi_nilai"`
	RataRataCapaian     float64 `gorm:"column:rata_rata_capaian"`
	TotalOnTrack        int64   `gorm:"column:total_on_track"`
	TotalWarning        int64   `gorm:"column:total_warning"`
	TotalOffTrack       int64   `gorm:"column:total_off_track"`
}

type programPerformanceRankingRow struct {
	ProgramID           uint64  `gorm:"column:program_id"`
	ProgramKode         string  `gorm:"column:program_kode"`
	ProgramNama         string  `gorm:"column:program_nama"`
	TotalTargetNilai    float64 `gorm:"column:total_target_nilai"`
	TotalRealisasiNilai float64 `gorm:"column:total_realisasi_nilai"`
	TotalIndikator      int64   `gorm:"column:total_indikator"`
	RataRataCapaian     float64 `gorm:"column:rata_rata_capaian"`
}

func NewHandler(cfg *config.Config) *Handler {
	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("performance handler unavailable: %v", err)
		return &Handler{ready: false, reason: "database connection unavailable"}
	}

	return &Handler{db: db, ready: true}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	performance := v1.Group("/performance")
	performance.GET("/informasi", h.GetInformasi)
	performance.GET("/informasi/latest", h.GetInformasiLatest)
	performance.POST("/informasi", h.CreateInformasi)
	performance.PUT("/informasi/:id", h.UpdateInformasi)
	performance.DELETE("/informasi/:id", h.DeleteInformasi)
	performance.GET("/target-realisasi", h.GetTargetRealisasi)
	performance.POST("/target-realisasi", h.CreateTargetRealisasi)
	performance.PUT("/target-realisasi/:id", h.UpdateTargetRealisasi)
	performance.GET("/calculate-achievement", h.CalculateAchievementPercentageByQuery)
	performance.POST("/calculate-achievement", h.CalculateAchievementPercentage)
	performance.GET("/dashboard-summary", h.GetDashboardSummary)
	performance.GET("/status-distribution", h.GetStatusDistribution)
	performance.GET("/statistics", h.GetPerformanceStatistics)
	performance.GET("/chart-target-vs-realisasi", h.GetChartTargetVsRealisasi)
	performance.GET("/yearly-summary", h.GetYearlyPerformanceSummary)
	performance.GET("/program-ranking", h.GetProgramPerformanceRanking)
}

func (h *Handler) GetInformasi(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	limit := 20
	if limitQ := strings.TrimSpace(c.Query("limit")); limitQ != "" {
		parsed, err := strconv.Atoi(limitQ)
		if err != nil || parsed < 1 {
			response.Error(c, http.StatusBadRequest, "invalid limit")
			return
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	q := h.db.WithContext(c.Request.Context()).Model(&database.Informasi{})
	if routeQ := strings.TrimSpace(c.Query("route")); routeQ != "" {
		route := normalizeInformasiRoute(routeQ)
		if route != "" {
			q = q.Where("pilihan_route_halaman_tujuan = ?", route)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count informasi")
		return
	}

	var items []database.Informasi
	if err := q.Order("tanggal_pembuatan DESC").Limit(limit).Find(&items).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load informasi")
		return
	}

	response.Success(c, gin.H{
		"total": total,
		"items": items,
	})
}

func (h *Handler) GetInformasiLatest(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	limit := 2
	if limitQ := strings.TrimSpace(c.Query("limit")); limitQ != "" {
		parsed, err := strconv.Atoi(limitQ)
		if err != nil || parsed < 1 {
			response.Error(c, http.StatusBadRequest, "invalid limit")
			return
		}
		if parsed > 10 {
			parsed = 10
		}
		limit = parsed
	}

	q := h.db.WithContext(c.Request.Context()).Model(&database.Informasi{})
	if routeQ := strings.TrimSpace(c.Query("route")); routeQ != "" {
		route := normalizeInformasiRoute(routeQ)
		if route != "" {
			q = q.Where("pilihan_route_halaman_tujuan = ?", route)
		}
	}

	var items []database.Informasi
	if err := q.
		Order("tanggal_pembuatan DESC").
		Limit(limit).
		Find(&items).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load latest informasi")
		return
	}

	var topikBaru *database.Informasi
	var topikLama *database.Informasi
	if len(items) > 0 {
		topikBaru = &items[0]
	}
	if len(items) > 1 {
		topikLama = &items[1]
	}

	response.Success(c, gin.H{
		"topik_baru": topikBaru,
		"topik_lama": topikLama,
		"items":      items,
	})
}

func (h *Handler) CreateInformasi(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	var req informasiUpsertRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}

	if req.Tahun < 1900 || req.Tahun > 3000 {
		response.Error(c, http.StatusBadRequest, "tahun is out of valid range")
		return
	}

	route := normalizeInformasiRoute(req.PilihanRouteHalamanTujuan)
	if route == "" {
		response.Error(c, http.StatusBadRequest, "invalid route")
		return
	}

	now := time.Now()
	item := database.Informasi{
		Informasi:                 strings.TrimSpace(req.Informasi),
		Tahun:                     req.Tahun,
		PilihanRouteHalamanTujuan: route,
		TanggalPembuatan:          now,
		TanggalUbah:               now,
	}

	if item.Informasi == "" {
		response.Error(c, http.StatusBadRequest, "informasi is required")
		return
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to create informasi")
		return
	}

	response.Success(c, item)
}

func (h *Handler) UpdateInformasi(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req informasiUpsertRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}

	if req.Tahun < 1900 || req.Tahun > 3000 {
		response.Error(c, http.StatusBadRequest, "tahun is out of valid range")
		return
	}

	route := normalizeInformasiRoute(req.PilihanRouteHalamanTujuan)
	if route == "" {
		response.Error(c, http.StatusBadRequest, "invalid route")
		return
	}

	var existing database.Informasi
	err = h.db.WithContext(c.Request.Context()).First(&existing, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "informasi not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get informasi")
		return
	}

	updates := map[string]any{
		"informasi":                    strings.TrimSpace(req.Informasi),
		"tahun":                        req.Tahun,
		"pilihan_route_halaman_tujuan": route,
		"tanggal_ubah":                 time.Now(),
	}

	if updates["informasi"] == "" {
		response.Error(c, http.StatusBadRequest, "informasi is required")
		return
	}

	if err := h.db.WithContext(c.Request.Context()).
		Model(&database.Informasi{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to update informasi")
		return
	}

	if err := h.db.WithContext(c.Request.Context()).First(&existing, "id = ?", id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load updated informasi")
		return
	}

	response.Success(c, existing)
}

func (h *Handler) DeleteInformasi(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	result := h.db.WithContext(c.Request.Context()).Delete(&database.Informasi{}, "id = ?", id)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, "failed to delete informasi")
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "informasi not found")
		return
	}

	response.Success(c, gin.H{"deleted_id": id})
}

func (h *Handler) GetStatusDistribution(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	tahun, hasTahun, err := parseOptionalInt(c.Query("tahun"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tahun")
		return
	}

	triwulan, hasTriwulan, err := parseOptionalInt(c.Query("triwulan"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid triwulan")
		return
	}
	if hasTriwulan && (triwulan < 1 || triwulan > 4) {
		response.Error(c, http.StatusBadRequest, "triwulan must be between 1 and 4")
		return
	}

	fallbackTriwulanExpr := strings.Join([]string{
		"CASE",
		"WHEN rr.triwulan IS NOT NULL THEN rr.triwulan",
		"WHEN rr.bulan BETWEEN 1 AND 3 THEN 1",
		"WHEN rr.bulan BETWEEN 4 AND 6 THEN 2",
		"WHEN rr.bulan BETWEEN 7 AND 9 THEN 3",
		"WHEN rr.bulan BETWEEN 10 AND 12 THEN 4",
		"ELSE 0",
		"END",
	}, " ")
	statusExpr := "CASE WHEN COALESCE(irk.target_tahunan, 0) <= 0 THEN 'UNKNOWN' WHEN ((rr.nilai_realisasi / irk.target_tahunan) * 100) >= 100 THEN 'ON_TRACK' WHEN ((rr.nilai_realisasi / irk.target_tahunan) * 100) >= 80 THEN 'WARNING' ELSE 'OFF_TRACK' END"
	capaianExpr := "CASE WHEN COALESCE(irk.target_tahunan, 0) = 0 THEN 0 ELSE (rr.nilai_realisasi / irk.target_tahunan) * 100 END"

	q := h.db.WithContext(c.Request.Context()).
		Table("realisasi_rencana_kerja rr").
		Joins("LEFT JOIN indikator_rencana_kerja irk ON irk.id = rr.indikator_rencana_kerja_id")
	if hasTahun {
		q = q.Where("rr.tahun = ?", tahun)
	}
	if hasTriwulan {
		q = q.Where("("+fallbackTriwulanExpr+") = ?", triwulan)
	}
	if query != "" {
		like := "%" + strings.ToLower(query) + "%"
		q = q.Where(
			"LOWER(irk.kode) LIKE ? OR LOWER(irk.nama) LIKE ? OR LOWER(rr.keterangan) LIKE ? OR CAST(rr.indikator_rencana_kerja_id AS CHAR) LIKE ? OR LOWER("+statusExpr+") LIKE ?",
			like, like, like, like, like,
		)
	}

	var agg struct {
		TotalData          int64   `gorm:"column:total_data"`
		TotalOnTrack       int64   `gorm:"column:total_on_track"`
		TotalWarning       int64   `gorm:"column:total_warning"`
		TotalOffTrack      int64   `gorm:"column:total_off_track"`
		TotalStatusUnknown int64   `gorm:"column:total_status_unknown"`
		RataRataCapaian    float64 `gorm:"column:rata_rata_capaian"`
	}

	if err := q.Select(strings.Join([]string{
		"COUNT(*) AS total_data",
		"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'ON_TRACK' THEN 1 ELSE 0 END), 0) AS total_on_track",
		"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'WARNING' THEN 1 ELSE 0 END), 0) AS total_warning",
		"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'OFF_TRACK' THEN 1 ELSE 0 END), 0) AS total_off_track",
		"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) NOT IN ('ON_TRACK', 'WARNING', 'OFF_TRACK') THEN 1 ELSE 0 END), 0) AS total_status_unknown",
		"COALESCE(AVG(" + capaianExpr + "), 0) AS rata_rata_capaian",
	}, ", ")).Scan(&agg).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to aggregate status distribution")
		return
	}

	percentageOf := func(value, total int64) float64 {
		if total <= 0 {
			return 0
		}
		return math.Round((float64(value)/float64(total))*100*100) / 100
	}

	response.Success(c, gin.H{
		"data_source": "realisasi_rencana_kerja",
		"filter": gin.H{
			"q":        queryValueOrText(query),
			"tahun":    queryValueOrAll(hasTahun, tahun),
			"triwulan": queryValueOrAll(hasTriwulan, triwulan),
		},
		"total_data":                  agg.TotalData,
		"total_status_on_track":       agg.TotalOnTrack,
		"total_status_warning":        agg.TotalWarning,
		"total_status_off_track":      agg.TotalOffTrack,
		"total_status_unknown":        agg.TotalStatusUnknown,
		"persentase_status_on_track":  percentageOf(agg.TotalOnTrack, agg.TotalData),
		"persentase_status_warning":   percentageOf(agg.TotalWarning, agg.TotalData),
		"persentase_status_off_track": percentageOf(agg.TotalOffTrack, agg.TotalData),
		"persentase_status_unknown":   percentageOf(agg.TotalStatusUnknown, agg.TotalData),
		"status_distribution": []gin.H{
			{"status": "ON_TRACK", "total": agg.TotalOnTrack, "persentase": percentageOf(agg.TotalOnTrack, agg.TotalData)},
			{"status": "WARNING", "total": agg.TotalWarning, "persentase": percentageOf(agg.TotalWarning, agg.TotalData)},
			{"status": "OFF_TRACK", "total": agg.TotalOffTrack, "persentase": percentageOf(agg.TotalOffTrack, agg.TotalData)},
			{"status": "UNKNOWN", "total": agg.TotalStatusUnknown, "persentase": percentageOf(agg.TotalStatusUnknown, agg.TotalData)},
		},
		"rata_rata_capaian_persen": math.Round(agg.RataRataCapaian*100) / 100,
	})
}

func (h *Handler) GetTargetRealisasi(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	query := strings.TrimSpace(c.Query("q"))

	page := 1
	if pageQ := strings.TrimSpace(c.Query("page")); pageQ != "" {
		parsed, err := strconv.Atoi(pageQ)
		if err != nil || parsed < 1 {
			response.Error(c, http.StatusBadRequest, "invalid page")
			return
		}
		page = parsed
	}

	limit := 10
	if limitQ := strings.TrimSpace(c.Query("limit")); limitQ != "" {
		parsed, err := strconv.Atoi(limitQ)
		if err != nil || parsed < 1 {
			response.Error(c, http.StatusBadRequest, "invalid limit")
			return
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	tahun, hasTahun, err := parseOptionalInt(c.Query("tahun"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tahun")
		return
	}

	capaianExpr := "COALESCE(tr.capaian_persen, CASE WHEN COALESCE(tr.target_nilai, 0) = 0 THEN 0 ELSE (tr.realisasi_nilai / tr.target_nilai) * 100 END)"
	statusExpr := "CASE WHEN TRIM(COALESCE(tr.status, '')) = '' THEN 'UNKNOWN' ELSE UPPER(TRIM(tr.status)) END"

	dbQuery := h.db.WithContext(c.Request.Context()).
		Table("target_dan_realisasi tr").
		Joins("LEFT JOIN indikator_rencana_kerja irk ON irk.id = tr.indikator_rencana_kerja_id")
	if hasTahun {
		dbQuery = dbQuery.Where("tr.tahun = ?", tahun)
	}

	if query != "" {
		like := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where(
			"CAST(tr.id AS CHAR) LIKE ? OR CAST(tr.indikator_rencana_kerja_id AS CHAR) LIKE ? OR CAST(tr.tahun AS CHAR) LIKE ? OR CAST(tr.triwulan AS CHAR) LIKE ? OR LOWER(irk.kode) LIKE ? OR LOWER(irk.nama) LIKE ? OR LOWER(tr.catatan) LIKE ? OR LOWER("+statusExpr+") LIKE ?",
			like, like, like, like, like, like, like, like,
		)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count target realisasi")
		return
	}

	dataSource := "target_dan_realisasi"

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * limit

	var items []database.TargetDanRealisasi
	if err := dbQuery.
		Select("tr.id, tr.indikator_rencana_kerja_id, tr.tahun, tr.triwulan, COALESCE(tr.target_nilai, 0) AS target_nilai, COALESCE(tr.realisasi_nilai, 0) AS realisasi_nilai, (" + capaianExpr + ") AS capaian_persen, (" + statusExpr + ") AS status, tr.diverifikasi_oleh, tr.tanggal_verifikasi, tr.catatan, tr.created_at, tr.updated_at").
		Order("tr.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load target realisasi")
		return
	}

	if total == 0 {
		fallbackTriwulanExpr := strings.Join([]string{
			"CASE",
			"WHEN rr.triwulan IS NOT NULL THEN rr.triwulan",
			"WHEN rr.bulan BETWEEN 1 AND 3 THEN 1",
			"WHEN rr.bulan BETWEEN 4 AND 6 THEN 2",
			"WHEN rr.bulan BETWEEN 7 AND 9 THEN 3",
			"WHEN rr.bulan BETWEEN 10 AND 12 THEN 4",
			"ELSE 0",
			"END",
		}, " ")
		fallbackCapaianExpr := "CASE WHEN COALESCE(irk.target_tahunan, 0) = 0 THEN 0 ELSE (rr.nilai_realisasi / irk.target_tahunan) * 100 END"
		fallbackStatusExpr := "CASE WHEN COALESCE(irk.target_tahunan, 0) <= 0 THEN 'UNKNOWN' WHEN ((rr.nilai_realisasi / irk.target_tahunan) * 100) >= 100 THEN 'ON_TRACK' WHEN ((rr.nilai_realisasi / irk.target_tahunan) * 100) >= 80 THEN 'WARNING' ELSE 'OFF_TRACK' END"

		fallbackQuery := h.db.WithContext(c.Request.Context()).
			Table("realisasi_rencana_kerja rr").
			Joins("LEFT JOIN indikator_rencana_kerja irk ON irk.id = rr.indikator_rencana_kerja_id")

		if hasTahun {
			fallbackQuery = fallbackQuery.Where("rr.tahun = ?", tahun)
		}

		if query != "" {
			like := "%" + strings.ToLower(query) + "%"
			fallbackQuery = fallbackQuery.Where(
				"CAST(rr.id AS CHAR) LIKE ? OR CAST(rr.indikator_rencana_kerja_id AS CHAR) LIKE ? OR CAST(rr.tahun AS CHAR) LIKE ? OR CAST(("+fallbackTriwulanExpr+") AS CHAR) LIKE ? OR LOWER(irk.kode) LIKE ? OR LOWER(irk.nama) LIKE ? OR LOWER(rr.keterangan) LIKE ? OR LOWER("+fallbackStatusExpr+") LIKE ?",
				like, like, like, like, like, like, like, like,
			)
		}

		if err := fallbackQuery.Count(&total).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to count target realisasi fallback")
			return
		}

		totalPages = int(math.Ceil(float64(total) / float64(limit)))
		if totalPages == 0 {
			totalPages = 1
		}
		if page > totalPages {
			page = totalPages
		}

		offset = (page - 1) * limit
		items = nil

		if err := fallbackQuery.
			Select("rr.id, rr.indikator_rencana_kerja_id, rr.tahun, (" + fallbackTriwulanExpr + ") AS triwulan, COALESCE(irk.target_tahunan, 0) AS target_nilai, COALESCE(rr.nilai_realisasi, 0) AS realisasi_nilai, (" + fallbackCapaianExpr + ") AS capaian_persen, (" + fallbackStatusExpr + ") AS status, NULL AS diverifikasi_oleh, NULL AS tanggal_verifikasi, rr.keterangan AS catatan, rr.created_at, rr.updated_at").
			Order("rr.id DESC").
			Limit(limit).
			Offset(offset).
			Find(&items).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to load target realisasi fallback")
			return
		}

		dataSource = "realisasi_rencana_kerja"
	}

	response.Success(c, gin.H{
		"module":      "Target dan Realisasi",
		"scope":       "Monitoring dan evaluasi capaian indikator",
		"status":      http.StatusText(http.StatusOK),
		"data_source": dataSource,
		"query":       query,
		"tahun":       queryValueOrAll(hasTahun, tahun),
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
		"items":       items,
	})
}

func (h *Handler) CreateTargetRealisasi(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	var req createTargetRealisasiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}

	if req.IndikatorRencanaKerjaID == 0 || req.Tahun == 0 || req.Triwulan == 0 {
		response.Error(c, http.StatusBadRequest, "indikator_rencana_kerja_id, tahun, and triwulan are required")
		return
	}
	if req.Triwulan < 1 || req.Triwulan > 4 {
		response.Error(c, http.StatusBadRequest, "triwulan must be between 1 and 4")
		return
	}

	item := database.TargetDanRealisasi{
		IndikatorRencanaKerjaID: req.IndikatorRencanaKerjaID,
		Tahun:                   req.Tahun,
		Triwulan:                req.Triwulan,
		TargetNilai:             req.TargetNilai,
		RealisasiNilai:          req.RealisasiNilai,
		Status:                  strings.TrimSpace(req.Status),
		Catatan:                 req.Catatan,
		DiverifikasiOleh:        req.DiverifikasiOleh,
	}

	if item.Status == "" {
		item.Status = "ON_TRACK"
	}

	if s := strings.TrimSpace(req.TanggalVerifikasi); s != "" {
		parsed, err := time.Parse(time.RFC3339, s)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "tanggal_verifikasi must be RFC3339 format")
			return
		}
		item.TanggalVerifikasi = &parsed
	}

	if err := h.db.WithContext(c.Request.Context()).
		Omit("capaian_persen").
		Create(&item).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to create target_realisasi")
		return
	}

	if err := h.db.WithContext(c.Request.Context()).First(&item, "id = ?", item.ID).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load created target_realisasi")
		return
	}

	response.Success(c, item)
}

func (h *Handler) UpdateTargetRealisasi(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req updateTargetRealisasiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}

	var item database.TargetDanRealisasi
	err = h.db.WithContext(c.Request.Context()).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "target_realisasi not found")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get target_realisasi")
		return
	}

	if req.TargetNilai != nil {
		item.TargetNilai = *req.TargetNilai
	}
	if req.RealisasiNilai != nil {
		item.RealisasiNilai = *req.RealisasiNilai
	}
	if s := strings.TrimSpace(req.Status); s != "" {
		item.Status = s
	}
	if req.Catatan != "" {
		item.Catatan = req.Catatan
	}
	if req.DiverifikasiOleh != nil {
		item.DiverifikasiOleh = req.DiverifikasiOleh
	}
	if s := strings.TrimSpace(req.TanggalVerifikasi); s != "" {
		parsed, err := time.Parse(time.RFC3339, s)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "tanggal_verifikasi must be RFC3339 format")
			return
		}
		item.TanggalVerifikasi = &parsed
	}

	updates := map[string]any{
		"target_nilai":       item.TargetNilai,
		"realisasi_nilai":    item.RealisasiNilai,
		"status":             item.Status,
		"catatan":            item.Catatan,
		"diverifikasi_oleh":  item.DiverifikasiOleh,
		"tanggal_verifikasi": item.TanggalVerifikasi,
	}

	if err := h.db.WithContext(c.Request.Context()).
		Model(&database.TargetDanRealisasi{}).
		Where("id = ?", id).
		Omit("capaian_persen", "created_at").
		Updates(updates).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to update target_realisasi")
		return
	}

	if err := h.db.WithContext(c.Request.Context()).First(&item, "id = ?", id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load updated target_realisasi")
		return
	}

	response.Success(c, item)
}

func (h *Handler) CalculateAchievementPercentage(c *gin.Context) {
	var req calculateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}
	h.respondAchievementCalculation(c, req.TargetNilai, req.RealisasiNilai)
}

func (h *Handler) CalculateAchievementPercentageByQuery(c *gin.Context) {
	target, err := strconv.ParseFloat(strings.TrimSpace(c.Query("target_nilai")), 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid target_nilai")
		return
	}

	realisasi, err := strconv.ParseFloat(strings.TrimSpace(c.Query("realisasi_nilai")), 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid realisasi_nilai")
		return
	}

	h.respondAchievementCalculation(c, target, realisasi)
}

func (h *Handler) respondAchievementCalculation(c *gin.Context, target, realisasi float64) {
	if target <= 0 {
		response.Error(c, http.StatusBadRequest, "target_nilai must be greater than 0")
		return
	}

	percentage := (realisasi / target) * 100
	percentage = math.Round(percentage*100) / 100

	response.Success(c, gin.H{
		"target_nilai":        target,
		"realisasi_nilai":     realisasi,
		"capaian_persen":      percentage,
		"perhitungan_formula": "(realisasi_nilai / target_nilai) * 100",
	})
}

// GetDashboardSummary returns a summary for the dashboard.
// This endpoint is explicitly allowed for all authenticated roles.
func (h *Handler) GetDashboardSummary(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	var totalProgram int64
	if err := h.db.WithContext(c.Request.Context()).Model(&database.Program{}).Count(&totalProgram).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count program")
		return
	}

	var totalKegiatan int64
	if err := h.db.WithContext(c.Request.Context()).Model(&database.Kegiatan{}).Count(&totalKegiatan).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count kegiatan")
		return
	}

	var totalSubKegiatan int64
	if err := h.db.WithContext(c.Request.Context()).Model(&database.SubKegiatan{}).Count(&totalSubKegiatan).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count sub_kegiatan")
		return
	}

	var totalRencanaKerja int64
	if err := h.db.WithContext(c.Request.Context()).Model(&database.RencanaKerja{}).Count(&totalRencanaKerja).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count rencana_kerja")
		return
	}

	var totalAnggaran float64
	if err := h.db.WithContext(c.Request.Context()).Model(&database.IndikatorRencanaKerja{}).
		Select("COALESCE(SUM(anggaran_tahunan), 0)").
		Scan(&totalAnggaran).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to sum anggaran")
		return
	}

	var totalRealisasiAnggaran float64
	if err := h.db.WithContext(c.Request.Context()).Model(&database.RealisasiRencanaKerja{}).
		Select("COALESCE(SUM(realisasi_anggaran), 0)").
		Scan(&totalRealisasiAnggaran).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to sum realisasi anggaran")
		return
	}

	persentaseRealisasiAnggaran := 0.0
	if totalAnggaran > 0 {
		persentaseRealisasiAnggaran = math.Round((totalRealisasiAnggaran/totalAnggaran)*100*100) / 100
	}

	response.Success(c, gin.H{
		"total_program":                 totalProgram,
		"total_kegiatan":                totalKegiatan,
		"total_sub_kegiatan":            totalSubKegiatan,
		"total_rencana_kerja":           totalRencanaKerja,
		"total_anggaran":                totalAnggaran,
		"total_realisasi_anggaran":      totalRealisasiAnggaran,
		"persentase_realisasi_anggaran": persentaseRealisasiAnggaran,
	})
}

func (h *Handler) GetPerformanceStatistics(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	tahun, hasTahun, err := parseOptionalInt(c.Query("tahun"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tahun")
		return
	}

	triwulan, hasTriwulan, err := parseOptionalInt(c.Query("triwulan"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid triwulan")
		return
	}
	if hasTriwulan && (triwulan < 1 || triwulan > 4) {
		response.Error(c, http.StatusBadRequest, "triwulan must be between 1 and 4")
		return
	}

	fallbackTriwulanExpr := strings.Join([]string{
		"CASE",
		"WHEN rr.triwulan IS NOT NULL THEN rr.triwulan",
		"WHEN rr.bulan BETWEEN 1 AND 3 THEN 1",
		"WHEN rr.bulan BETWEEN 4 AND 6 THEN 2",
		"WHEN rr.bulan BETWEEN 7 AND 9 THEN 3",
		"WHEN rr.bulan BETWEEN 10 AND 12 THEN 4",
		"ELSE 0",
		"END",
	}, " ")
	statusExpr := "CASE WHEN COALESCE(irk.target_tahunan, 0) <= 0 THEN 'UNKNOWN' WHEN ((rr.nilai_realisasi / irk.target_tahunan) * 100) >= 100 THEN 'ON_TRACK' WHEN ((rr.nilai_realisasi / irk.target_tahunan) * 100) >= 80 THEN 'WARNING' ELSE 'OFF_TRACK' END"
	capaianExpr := "CASE WHEN COALESCE(irk.target_tahunan, 0) = 0 THEN 0 ELSE (rr.nilai_realisasi / irk.target_tahunan) * 100 END"

	buildTargetQuery := func() *gorm.DB {
		q := h.db.WithContext(c.Request.Context()).
			Table("realisasi_rencana_kerja rr").
			Joins("LEFT JOIN indikator_rencana_kerja irk ON irk.id = rr.indikator_rencana_kerja_id")
		if hasTahun {
			q = q.Where("rr.tahun = ?", tahun)
		}
		if hasTriwulan {
			q = q.Where("("+fallbackTriwulanExpr+") = ?", triwulan)
		}
		if query != "" {
			like := "%" + strings.ToLower(query) + "%"
			q = q.Where(
				"LOWER(irk.kode) LIKE ? OR LOWER(irk.nama) LIKE ? OR LOWER(rr.keterangan) LIKE ? OR CAST(rr.indikator_rencana_kerja_id AS CHAR) LIKE ? OR LOWER("+statusExpr+") LIKE ?",
				like, like, like, like,
				like,
			)
		}
		return q
	}

	var totalIndikator int64
	if err := buildTargetQuery().Distinct("rr.indikator_rencana_kerja_id").Count(&totalIndikator).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count indikator")
		return
	}

	var capaianAgg struct {
		TotalData           int64 `gorm:"column:total_data"`
		TotalOnTrack        int64 `gorm:"column:total_on_track"`
		TotalWarning        int64 `gorm:"column:total_warning"`
		TotalOffTrack       int64 `gorm:"column:total_off_track"`
		TotalStatusUnknown  int64 `gorm:"column:total_status_unknown"`
		TotalTargetNilai    float64
		TotalRealisasiNilai float64
		RataRataCapaian     float64
	}
	if err := buildTargetQuery().
		Select(strings.Join([]string{
			"COUNT(*) AS total_data",
			"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'ON_TRACK' THEN 1 ELSE 0 END), 0) AS total_on_track",
			"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'WARNING' THEN 1 ELSE 0 END), 0) AS total_warning",
			"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'OFF_TRACK' THEN 1 ELSE 0 END), 0) AS total_off_track",
			"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) NOT IN ('ON_TRACK', 'WARNING', 'OFF_TRACK') THEN 1 ELSE 0 END), 0) AS total_status_unknown",
			"COALESCE(SUM(irk.target_tahunan), 0) AS total_target_nilai",
			"COALESCE(SUM(rr.nilai_realisasi), 0) AS total_realisasi_nilai",
			"COALESCE(AVG(" + capaianExpr + "), 0) AS rata_rata_capaian",
		}, ", ")).
		Scan(&capaianAgg).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to aggregate capaian")
		return
	}

	percentageOf := func(value, total int64) float64 {
		if total <= 0 {
			return 0
		}
		return math.Round((float64(value)/float64(total))*100*100) / 100
	}

	var triwulanItems []gin.H
	if !hasTriwulan {
		var triwulanRows []struct {
			Triwulan       int     `gorm:"column:triwulan"`
			TotalData      int64   `gorm:"column:total_data"`
			TotalOnTrack   int64   `gorm:"column:total_on_track"`
			TotalWarning   int64   `gorm:"column:total_warning"`
			TotalOffTrack  int64   `gorm:"column:total_off_track"`
			TotalUnknown   int64   `gorm:"column:total_unknown"`
			AvgCapaian     float64 `gorm:"column:rata_rata_capaian"`
			TotalTarget    float64 `gorm:"column:total_target_nilai"`
			TotalRealisasi float64 `gorm:"column:total_realisasi_nilai"`
		}

		if err := buildTargetQuery().
			Select(strings.Join([]string{
				"(" + fallbackTriwulanExpr + ") AS triwulan",
				"COUNT(*) AS total_data",
				"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'ON_TRACK' THEN 1 ELSE 0 END), 0) AS total_on_track",
				"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'WARNING' THEN 1 ELSE 0 END), 0) AS total_warning",
				"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) = 'OFF_TRACK' THEN 1 ELSE 0 END), 0) AS total_off_track",
				"COALESCE(SUM(CASE WHEN UPPER(TRIM(" + statusExpr + ")) NOT IN ('ON_TRACK', 'WARNING', 'OFF_TRACK') THEN 1 ELSE 0 END), 0) AS total_unknown",
				"COALESCE(AVG(" + capaianExpr + "), 0) AS rata_rata_capaian",
				"COALESCE(SUM(irk.target_tahunan), 0) AS total_target_nilai",
				"COALESCE(SUM(rr.nilai_realisasi), 0) AS total_realisasi_nilai",
			}, ", ")).
			Group("(" + fallbackTriwulanExpr + ")").
			Order("(" + fallbackTriwulanExpr + ") ASC").
			Scan(&triwulanRows).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to aggregate triwulan statistics")
			return
		}

		triwulanItems = make([]gin.H, 0, len(triwulanRows))
		for _, row := range triwulanRows {
			if row.Triwulan < 1 || row.Triwulan > 4 {
				continue
			}
			persentaseRealisasiTargetTriwulan := 0.0
			if row.TotalTarget > 0 {
				persentaseRealisasiTargetTriwulan = math.Round((row.TotalRealisasi/row.TotalTarget)*100*100) / 100
			}

			triwulanItems = append(triwulanItems, gin.H{
				"triwulan":                    row.Triwulan,
				"total_data":                  row.TotalData,
				"total_status_on_track":       row.TotalOnTrack,
				"total_status_warning":        row.TotalWarning,
				"total_status_off_track":      row.TotalOffTrack,
				"total_status_unknown":        row.TotalUnknown,
				"rata_rata_capaian_persen":    math.Round(row.AvgCapaian*100) / 100,
				"persentase_realisasi_target": persentaseRealisasiTargetTriwulan,
			})
		}
	}

	persentaseRealisasiTarget := 0.0
	if capaianAgg.TotalTargetNilai > 0 {
		persentaseRealisasiTarget = math.Round((capaianAgg.TotalRealisasiNilai/capaianAgg.TotalTargetNilai)*100*100) / 100
	}

	response.Success(c, gin.H{
		"data_source": "realisasi_rencana_kerja",
		"filter": gin.H{
			"q":        queryValueOrText(query),
			"tahun":    queryValueOrAll(hasTahun, tahun),
			"triwulan": queryValueOrAll(hasTriwulan, triwulan),
		},
		"total_data":                  capaianAgg.TotalData,
		"total_indikator":             totalIndikator,
		"total_status_on_track":       capaianAgg.TotalOnTrack,
		"total_status_warning":        capaianAgg.TotalWarning,
		"total_status_off_track":      capaianAgg.TotalOffTrack,
		"total_status_unknown":        capaianAgg.TotalStatusUnknown,
		"persentase_status_on_track":  percentageOf(capaianAgg.TotalOnTrack, capaianAgg.TotalData),
		"persentase_status_warning":   percentageOf(capaianAgg.TotalWarning, capaianAgg.TotalData),
		"persentase_status_off_track": percentageOf(capaianAgg.TotalOffTrack, capaianAgg.TotalData),
		"persentase_status_unknown":   percentageOf(capaianAgg.TotalStatusUnknown, capaianAgg.TotalData),
		"status_distribution": []gin.H{
			{"status": "ON_TRACK", "total": capaianAgg.TotalOnTrack, "persentase": percentageOf(capaianAgg.TotalOnTrack, capaianAgg.TotalData)},
			{"status": "WARNING", "total": capaianAgg.TotalWarning, "persentase": percentageOf(capaianAgg.TotalWarning, capaianAgg.TotalData)},
			{"status": "OFF_TRACK", "total": capaianAgg.TotalOffTrack, "persentase": percentageOf(capaianAgg.TotalOffTrack, capaianAgg.TotalData)},
			{"status": "UNKNOWN", "total": capaianAgg.TotalStatusUnknown, "persentase": percentageOf(capaianAgg.TotalStatusUnknown, capaianAgg.TotalData)},
		},
		"total_target_nilai":          capaianAgg.TotalTargetNilai,
		"total_realisasi_nilai":       capaianAgg.TotalRealisasiNilai,
		"rata_rata_capaian_persen":    math.Round(capaianAgg.RataRataCapaian*100) / 100,
		"persentase_realisasi_target": persentaseRealisasiTarget,
		"items_triwulan":              triwulanItems,
	})
}

func (h *Handler) GetChartTargetVsRealisasi(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	tahun, hasTahun, err := parseOptionalInt(c.Query("tahun"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tahun")
		return
	}

	q := h.db.WithContext(c.Request.Context()).
		Table("target_dan_realisasi tr")
	if hasTahun {
		q = q.Where("tr.tahun = ?", tahun)
	}

	var rows []chartTargetVsRealisasiRow
	if err := q.
		Select("tr.triwulan AS triwulan, COALESCE(SUM(tr.target_nilai), 0) AS total_target, COALESCE(SUM(tr.realisasi_nilai), 0) AS total_realisasi").
		Where("tr.triwulan BETWEEN 1 AND 4").
		Group("tr.triwulan").
		Order("tr.triwulan ASC").
		Scan(&rows).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to aggregate chart data")
		return
	}

	categories := []string{"T1", "T2", "T3", "T4"}
	seriesTarget := []float64{0, 0, 0, 0}
	seriesRealisasi := []float64{0, 0, 0, 0}

	for _, row := range rows {
		if row.Triwulan < 1 || row.Triwulan > 4 {
			continue
		}
		idx := row.Triwulan - 1
		seriesTarget[idx] = math.Round(row.TotalTarget*100) / 100
		seriesRealisasi[idx] = math.Round(row.TotalRealisasi*100) / 100
	}

	response.Success(c, gin.H{
		"filter": gin.H{
			"tahun": queryValueOrAll(hasTahun, tahun),
		},
		"categories": categories,
		"series": gin.H{
			"target":    seriesTarget,
			"realisasi": seriesRealisasi,
		},
	})
}

func (h *Handler) GetYearlyPerformanceSummary(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	tahunStart, hasTahunStart, err := parseOptionalInt(c.Query("tahun_start"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tahun_start")
		return
	}

	tahunEnd, hasTahunEnd, err := parseOptionalInt(c.Query("tahun_end"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tahun_end")
		return
	}

	if hasTahunStart && hasTahunEnd && tahunStart > tahunEnd {
		response.Error(c, http.StatusBadRequest, "tahun_start must be less than or equal to tahun_end")
		return
	}

	dataSource := "target_dan_realisasi"

	q := h.db.WithContext(c.Request.Context()).Model(&database.TargetDanRealisasi{})
	if hasTahunStart {
		q = q.Where("tahun >= ?", tahunStart)
	}
	if hasTahunEnd {
		q = q.Where("tahun <= ?", tahunEnd)
	}

	var rows []yearlyPerformanceSummaryRow
	if err := q.
		Select("tahun, COUNT(*) AS total_data, COUNT(DISTINCT indikator_rencana_kerja_id) AS total_indikator, COALESCE(SUM(target_nilai), 0) AS total_target_nilai, COALESCE(SUM(realisasi_nilai), 0) AS total_realisasi_nilai, COALESCE(AVG(capaian_persen), 0) AS rata_rata_capaian, COALESCE(SUM(CASE WHEN status = 'ON_TRACK' THEN 1 ELSE 0 END), 0) AS total_on_track, COALESCE(SUM(CASE WHEN status = 'WARNING' THEN 1 ELSE 0 END), 0) AS total_warning, COALESCE(SUM(CASE WHEN status = 'OFF_TRACK' THEN 1 ELSE 0 END), 0) AS total_off_track").
		Group("tahun").
		Order("tahun ASC").
		Scan(&rows).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to aggregate yearly summary")
		return
	}

	// Fallback: when target_dan_realisasi has no data for the given period,
	// aggregate from realisasi_rencana_kerja; target proxied by irk.target_tahunan.
	if len(rows) == 0 {
		dataSource = "realisasi_rencana_kerja"
		q2 := h.db.WithContext(c.Request.Context()).
			Table("realisasi_rencana_kerja rrk").
			Joins("JOIN indikator_rencana_kerja irk ON irk.id = rrk.indikator_rencana_kerja_id")
		if hasTahunStart {
			q2 = q2.Where("rrk.tahun >= ?", tahunStart)
		}
		if hasTahunEnd {
			q2 = q2.Where("rrk.tahun <= ?", tahunEnd)
		}
		if err := q2.
			Select(strings.Join([]string{
				"rrk.tahun AS tahun",
				"COUNT(*) AS total_data",
				"COUNT(DISTINCT rrk.indikator_rencana_kerja_id) AS total_indikator",
				"COALESCE(SUM(irk.target_tahunan), 0) AS total_target_nilai",
				"COALESCE(SUM(rrk.nilai_realisasi), 0) AS total_realisasi_nilai",
				"COALESCE(AVG(CASE WHEN irk.target_tahunan > 0 THEN (rrk.nilai_realisasi / irk.target_tahunan * 100) ELSE 0 END), 0) AS rata_rata_capaian",
				"0 AS total_on_track",
				"0 AS total_warning",
				"0 AS total_off_track",
			}, ", ")).
			Group("rrk.tahun").
			Order("rrk.tahun ASC").
			Scan(&rows).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to aggregate yearly summary (fallback)")
			return
		}
	}

	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		persentaseRealisasiTarget := 0.0
		if row.TotalTargetNilai > 0 {
			persentaseRealisasiTarget = math.Round((row.TotalRealisasiNilai/row.TotalTargetNilai)*100*100) / 100
		}

		items = append(items, gin.H{
			"tahun":                       row.Tahun,
			"total_data":                  row.TotalData,
			"total_indikator":             row.TotalIndikator,
			"total_target_nilai":          math.Round(row.TotalTargetNilai*100) / 100,
			"total_realisasi_nilai":       math.Round(row.TotalRealisasiNilai*100) / 100,
			"rata_rata_capaian_persen":    math.Round(row.RataRataCapaian*100) / 100,
			"persentase_realisasi_target": persentaseRealisasiTarget,
			"total_status_on_track":       row.TotalOnTrack,
			"total_status_warning":        row.TotalWarning,
			"total_status_off_track":      row.TotalOffTrack,
		})
	}

	response.Success(c, gin.H{
		"filter": gin.H{
			"tahun_start": queryValueOrAll(hasTahunStart, tahunStart),
			"tahun_end":   queryValueOrAll(hasTahunEnd, tahunEnd),
		},
		"total_tahun": len(items),
		"data_source": dataSource,
		"items":       items,
	})
}

func (h *Handler) GetProgramPerformanceRanking(c *gin.Context) {
	if !h.ensureReady(c) {
		return
	}

	query := strings.TrimSpace(c.Query("q"))

	page := 1
	if pageQ := strings.TrimSpace(c.Query("page")); pageQ != "" {
		parsed, err := strconv.Atoi(pageQ)
		if err != nil || parsed < 1 {
			response.Error(c, http.StatusBadRequest, "invalid page")
			return
		}
		page = parsed
	}

	tahun, hasTahun, err := parseOptionalInt(c.Query("tahun"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tahun")
		return
	}

	triwulan, hasTriwulan, err := parseOptionalInt(c.Query("triwulan"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid triwulan")
		return
	}
	if hasTriwulan && (triwulan < 1 || triwulan > 4) {
		response.Error(c, http.StatusBadRequest, "triwulan must be between 1 and 4")
		return
	}

	limit := 10
	if limitQ := strings.TrimSpace(c.Query("limit")); limitQ != "" {
		limitParsed, err := strconv.Atoi(limitQ)
		if err != nil || limitParsed <= 0 {
			response.Error(c, http.StatusBadRequest, "invalid limit")
			return
		}
		if limitParsed > 100 {
			limitParsed = 100
		}
		limit = limitParsed
	}

	dataSource := "target_dan_realisasi"

	q := h.db.WithContext(c.Request.Context()).Table("target_dan_realisasi tr").
		Joins("JOIN indikator_rencana_kerja irk ON irk.id = tr.indikator_rencana_kerja_id").
		Joins("LEFT JOIN rencana_kerja rk ON rk.id = irk.rencana_kerja_id").
		Joins("LEFT JOIN indikator_sub_kegiatan isk ON isk.id = rk.indikator_sub_kegiatan_id").
		Joins("LEFT JOIN sub_kegiatan sk ON sk.id = isk.sub_kegiatan_id").
		Joins("LEFT JOIN kegiatan k ON k.id = sk.kegiatan_id").
		Joins("LEFT JOIN program p ON p.id = k.program_id")

	if hasTahun {
		q = q.Where("tr.tahun = ?", tahun)
	}
	if hasTriwulan {
		q = q.Where("tr.triwulan = ?", triwulan)
	}
	if query != "" {
		like := "%" + strings.ToLower(query) + "%"
		q = q.Where(
			"LOWER(COALESCE(p.kode, '')) LIKE ? OR LOWER(COALESCE(p.nama, '')) LIKE ?",
			like,
			like,
		)
	}

	baseAgg := q.Select(strings.Join([]string{
		"COALESCE(p.id, 0) AS program_id",
		"COALESCE(NULLIF(TRIM(p.kode), ''), '-') AS program_kode",
		"COALESCE(NULLIF(TRIM(p.nama), ''), 'Program Belum Terpetakan') AS program_nama",
		"COALESCE(SUM(tr.target_nilai), 0) AS total_target_nilai",
		"COALESCE(SUM(tr.realisasi_nilai), 0) AS total_realisasi_nilai",
		"COUNT(DISTINCT tr.indikator_rencana_kerja_id) AS total_indikator",
		"COALESCE(AVG(tr.capaian_persen), 0) AS rata_rata_capaian",
	}, ", ")).
		Group(strings.Join([]string{
			"COALESCE(p.id, 0)",
			"COALESCE(NULLIF(TRIM(p.kode), ''), '-')",
			"COALESCE(NULLIF(TRIM(p.nama), ''), 'Program Belum Terpetakan')",
		}, ", "))

	var total int64
	if err := h.db.WithContext(c.Request.Context()).
		Table("(?) AS program_rankings", baseAgg).
		Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count program ranking")
		return
	}

	// Fallback: when target_dan_realisasi has no data, aggregate from realisasi_rencana_kerja.
	// target_nilai is proxied by irk.target_tahunan; capaian is derived on-the-fly.
	if total == 0 {
		dataSource = "realisasi_rencana_kerja"
		q2 := h.db.WithContext(c.Request.Context()).Table("realisasi_rencana_kerja rrk").
			Joins("JOIN indikator_rencana_kerja irk ON irk.id = rrk.indikator_rencana_kerja_id").
			Joins("LEFT JOIN rencana_kerja rk ON rk.id = irk.rencana_kerja_id").
			Joins("LEFT JOIN indikator_sub_kegiatan isk ON isk.id = rk.indikator_sub_kegiatan_id").
			Joins("LEFT JOIN sub_kegiatan sk ON sk.id = isk.sub_kegiatan_id").
			Joins("LEFT JOIN kegiatan k ON k.id = sk.kegiatan_id").
			Joins("LEFT JOIN program p ON p.id = k.program_id")

		if hasTahun {
			q2 = q2.Where("rrk.tahun = ?", tahun)
		}
		if hasTriwulan {
			q2 = q2.Where("rrk.triwulan = ?", triwulan)
		}
		if query != "" {
			like := "%" + strings.ToLower(query) + "%"
			q2 = q2.Where(
				"LOWER(COALESCE(p.kode, '')) LIKE ? OR LOWER(COALESCE(p.nama, '')) LIKE ?",
				like,
				like,
			)
		}

		baseAgg = q2.Select(strings.Join([]string{
			"COALESCE(p.id, 0) AS program_id",
			"COALESCE(NULLIF(TRIM(p.kode), ''), '-') AS program_kode",
			"COALESCE(NULLIF(TRIM(p.nama), ''), 'Program Belum Terpetakan') AS program_nama",
			"COALESCE(SUM(irk.target_tahunan), 0) AS total_target_nilai",
			"COALESCE(SUM(rrk.nilai_realisasi), 0) AS total_realisasi_nilai",
			"COUNT(DISTINCT rrk.indikator_rencana_kerja_id) AS total_indikator",
			"COALESCE(AVG(CASE WHEN irk.target_tahunan > 0 THEN (rrk.nilai_realisasi / irk.target_tahunan * 100) ELSE 0 END), 0) AS rata_rata_capaian",
		}, ", ")).
			Group(strings.Join([]string{
				"COALESCE(p.id, 0)",
				"COALESCE(NULLIF(TRIM(p.kode), ''), '-')",
				"COALESCE(NULLIF(TRIM(p.nama), ''), 'Program Belum Terpetakan')",
			}, ", "))

		if err := h.db.WithContext(c.Request.Context()).
			Table("(?) AS program_rankings", baseAgg).
			Count(&total).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to count program ranking (fallback)")
			return
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * limit

	var rows []programPerformanceRankingRow
	if err := h.db.WithContext(c.Request.Context()).
		Table("(?) AS program_rankings", baseAgg).
		Order("(COALESCE(total_realisasi_nilai, 0) / NULLIF(COALESCE(total_target_nilai, 0), 0)) DESC, COALESCE(total_realisasi_nilai, 0) DESC, program_nama ASC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to aggregate program ranking")
		return
	}

	items := make([]gin.H, 0, len(rows))
	for i, row := range rows {
		persentaseRealisasiTarget := 0.0
		if row.TotalTargetNilai > 0 {
			persentaseRealisasiTarget = math.Round((row.TotalRealisasiNilai/row.TotalTargetNilai)*100*100) / 100
		}

		items = append(items, gin.H{
			"rank":                        offset + i + 1,
			"program_id":                  row.ProgramID,
			"program_kode":                row.ProgramKode,
			"program_nama":                row.ProgramNama,
			"total_indikator":             row.TotalIndikator,
			"total_target_nilai":          math.Round(row.TotalTargetNilai*100) / 100,
			"total_realisasi_nilai":       math.Round(row.TotalRealisasiNilai*100) / 100,
			"rata_rata_capaian_persen":    math.Round(row.RataRataCapaian*100) / 100,
			"persentase_realisasi_target": persentaseRealisasiTarget,
		})
	}

	response.Success(c, gin.H{
		"filter": gin.H{
			"q":        queryValueOrText(query),
			"tahun":    queryValueOrAll(hasTahun, tahun),
			"triwulan": queryValueOrAll(hasTriwulan, triwulan),
			"page":     page,
			"limit":    limit,
		},
		"page":          page,
		"limit":         limit,
		"total":         total,
		"total_pages":   totalPages,
		"total_program": total,
		"data_source":   dataSource,
		"items":         items,
	})
}

func parseOptionalInt(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false, nil
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, false, err
	}
	return v, true, nil
}

func queryValueOrAll(hasValue bool, value int) any {
	if !hasValue {
		return "all"
	}
	return value
}

func queryValueOrText(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "all"
	}
	return trimmed
}

func normalizeInformasiRoute(route string) string {
	normalized := strings.TrimRight(strings.TrimSpace(route), "/")
	if normalized == "" {
		return ""
	}

	// List of recognized routes that have specific information topics
	recognized := map[string]bool{
		"/dashboard":               true,
		"/rencana-kerja":           true,
		"/dokumen_pdf":             true,
		"/target-evaluasi":         true,
		"/target-realisasi":        true,
		"/informasi":               true,
		"/visi":                    true,
		"/misi":                    true,
		"/tujuan":                  true,
		"/sasaran":                 true,
		"/program":                 true,
		"/kegiatan":                true,
		"/sub-kegiatan":            true,
		"/pagu-sub-kegiatan":       true,
		"/unit-pengusul":           true,
		"/indikator-tujuan":        true,
		"/indikator-sasaran":       true,
		"/indikator-program":       true,
		"/indikator-kegiatan":      true,
		"/indikator-sub-kegiatan":  true,
		"/manajemen-user":          true,
		"/rencana-kerja-spa":       true,
	}

	if recognized[normalized] {
		if normalized == "/target-realisasi" {
			return "/target-evaluasi"
		}
		return normalized
	}

	// If not recognized specifically, still return it if it starts with /
	// This allows future-proofing and doesn't break if a new page is added without updating this list.
	if strings.HasPrefix(normalized, "/") {
		return normalized
	}

	return ""
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
