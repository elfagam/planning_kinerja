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

func NewHandler(cfg config.Config) *Handler {
	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("performance handler unavailable: %v", err)
		return &Handler{ready: false, reason: "database connection unavailable"}
	}

	return &Handler{db: db, ready: true}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	performance := v1.Group("/performance")
	performance.GET("/target-realisasi", h.GetTargetRealisasi)
	performance.POST("/target-realisasi", h.CreateTargetRealisasi)
	performance.PUT("/target-realisasi/:id", h.UpdateTargetRealisasi)
	performance.GET("/calculate-achievement", h.CalculateAchievementPercentageByQuery)
	performance.POST("/calculate-achievement", h.CalculateAchievementPercentage)
	performance.GET("/dashboard-summary", h.GetDashboardSummary)
	performance.GET("/statistics", h.GetPerformanceStatistics)
	performance.GET("/chart-target-vs-realisasi", h.GetChartTargetVsRealisasi)
	performance.GET("/yearly-summary", h.GetYearlyPerformanceSummary)
	performance.GET("/program-ranking", h.GetProgramPerformanceRanking)
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

	dbQuery := h.db.WithContext(c.Request.Context()).Model(&database.TargetDanRealisasi{})
	if hasTahun {
		dbQuery = dbQuery.Where("tahun = ?", tahun)
	}

	if query != "" {
		like := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where(
			"LOWER(status) LIKE ? OR CAST(id AS CHAR) LIKE ? OR CAST(indikator_rencana_kerja_id AS CHAR) LIKE ? OR CAST(tahun AS CHAR) LIKE ? OR CAST(triwulan AS CHAR) LIKE ?",
			like, like, like, like, like,
		)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count target realisasi")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * limit

	var items []database.TargetDanRealisasi
	if err := dbQuery.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load target realisasi")
		return
	}

	response.Success(c, gin.H{
		"module":      "Target dan Realisasi",
		"scope":       "Monitoring dan evaluasi capaian indikator",
		"status":      http.StatusText(http.StatusOK),
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

	if item.TargetNilai > 0 {
		item.CapaianPersen = (item.RealisasiNilai / item.TargetNilai) * 100
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&item).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to create target_realisasi")
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

	if item.TargetNilai > 0 {
		item.CapaianPersen = (item.RealisasiNilai / item.TargetNilai) * 100
	}

	if err := h.db.WithContext(c.Request.Context()).Save(&item).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "failed to update target_realisasi")
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

	buildTargetQuery := func() *gorm.DB {
		q := h.db.WithContext(c.Request.Context()).Model(&database.TargetDanRealisasi{})
		if hasTahun {
			q = q.Where("tahun = ?", tahun)
		}
		if hasTriwulan {
			q = q.Where("triwulan = ?", triwulan)
		}
		return q
	}

	var totalData int64
	if err := buildTargetQuery().Count(&totalData).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count target_realisasi")
		return
	}

	var totalOnTrack int64
	if err := buildTargetQuery().Where("status = ?", "ON_TRACK").Count(&totalOnTrack).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count on_track status")
		return
	}

	var totalWarning int64
	if err := buildTargetQuery().Where("status = ?", "WARNING").Count(&totalWarning).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count warning status")
		return
	}

	var totalOffTrack int64
	if err := buildTargetQuery().Where("status = ?", "OFF_TRACK").Count(&totalOffTrack).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count off_track status")
		return
	}

	var totalIndikator int64
	if err := buildTargetQuery().Distinct("indikator_rencana_kerja_id").Count(&totalIndikator).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count indikator")
		return
	}

	var capaianAgg struct {
		TotalTargetNilai    float64
		TotalRealisasiNilai float64
		RataRataCapaian     float64
	}
	if err := buildTargetQuery().
		Select("COALESCE(SUM(target_nilai), 0) AS total_target_nilai, COALESCE(SUM(realisasi_nilai), 0) AS total_realisasi_nilai, COALESCE(AVG(capaian_persen), 0) AS rata_rata_capaian").
		Scan(&capaianAgg).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to aggregate capaian")
		return
	}

	persentaseRealisasiTarget := 0.0
	if capaianAgg.TotalTargetNilai > 0 {
		persentaseRealisasiTarget = math.Round((capaianAgg.TotalRealisasiNilai/capaianAgg.TotalTargetNilai)*100*100) / 100
	}

	response.Success(c, gin.H{
		"filter": gin.H{
			"tahun":    queryValueOrAll(hasTahun, tahun),
			"triwulan": queryValueOrAll(hasTriwulan, triwulan),
		},
		"total_data":                  totalData,
		"total_indikator":             totalIndikator,
		"total_status_on_track":       totalOnTrack,
		"total_status_warning":        totalWarning,
		"total_status_off_track":      totalOffTrack,
		"total_target_nilai":          capaianAgg.TotalTargetNilai,
		"total_realisasi_nilai":       capaianAgg.TotalRealisasiNilai,
		"rata_rata_capaian_persen":    math.Round(capaianAgg.RataRataCapaian*100) / 100,
		"persentase_realisasi_target": persentaseRealisasiTarget,
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

	q := h.db.WithContext(c.Request.Context()).Model(&database.TargetDanRealisasi{})
	if hasTahun {
		q = q.Where("tahun = ?", tahun)
	}

	var rows []chartTargetVsRealisasiRow
	if err := q.
		Select("tahun, triwulan, COALESCE(SUM(target_nilai), 0) AS total_target, COALESCE(SUM(realisasi_nilai), 0) AS total_realisasi").
		Group("tahun, triwulan").
		Order("tahun ASC, triwulan ASC").
		Scan(&rows).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to aggregate chart data")
		return
	}

	categories := make([]string, 0, len(rows))
	seriesTarget := make([]float64, 0, len(rows))
	seriesRealisasi := make([]float64, 0, len(rows))

	for _, row := range rows {
		categories = append(categories, strconv.Itoa(row.Tahun)+"-T"+strconv.Itoa(row.Triwulan))
		seriesTarget = append(seriesTarget, math.Round(row.TotalTarget*100)/100)
		seriesRealisasi = append(seriesRealisasi, math.Round(row.TotalRealisasi*100)/100)
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
		"items":       items,
	})
}

func (h *Handler) GetProgramPerformanceRanking(c *gin.Context) {
	if !h.ensureReady(c) {
		return
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

	q := h.db.WithContext(c.Request.Context()).Table("target_dan_realisasi tr").
		Joins("JOIN indikator_rencana_kerja irk ON irk.id = tr.indikator_rencana_kerja_id").
		Joins("JOIN rencana_kerja rk ON rk.id = irk.rencana_kerja_id").
		Joins("JOIN indikator_sub_kegiatan isk ON isk.id = rk.indikator_sub_kegiatan_id").
		Joins("JOIN sub_kegiatan sk ON sk.id = isk.sub_kegiatan_id").
		Joins("JOIN kegiatan k ON k.id = sk.kegiatan_id").
		Joins("JOIN program p ON p.id = k.program_id")

	if hasTahun {
		q = q.Where("tr.tahun = ?", tahun)
	}
	if hasTriwulan {
		q = q.Where("tr.triwulan = ?", triwulan)
	}

	var rows []programPerformanceRankingRow
	if err := q.
		Select("p.id AS program_id, p.kode AS program_kode, p.nama AS program_nama, COALESCE(SUM(tr.target_nilai), 0) AS total_target_nilai, COALESCE(SUM(tr.realisasi_nilai), 0) AS total_realisasi_nilai, COUNT(DISTINCT tr.indikator_rencana_kerja_id) AS total_indikator, COALESCE(AVG(tr.capaian_persen), 0) AS rata_rata_capaian").
		Group("p.id, p.kode, p.nama").
		Order("(COALESCE(SUM(tr.realisasi_nilai), 0) / NULLIF(COALESCE(SUM(tr.target_nilai), 0), 0)) DESC, COALESCE(SUM(tr.realisasi_nilai), 0) DESC, p.nama ASC").
		Limit(limit).
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
			"rank":                        i + 1,
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
			"tahun":    queryValueOrAll(hasTahun, tahun),
			"triwulan": queryValueOrAll(hasTriwulan, triwulan),
			"limit":    limit,
		},
		"total_program": len(items),
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
