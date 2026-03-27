package http

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/modules/renja/domain"
	renjarepo "e-plan-ai/internal/modules/renja/repository"
	"e-plan-ai/internal/modules/renja/usecase"
	shareddb "e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service *usecase.Service
	storage string
}

type actionRequest struct {
	ActorID int64  `json:"actor_id"`
	Reason  string `json:"reason"`
}

func NewHandler(cfg *config.Config) *Handler {
	db, err := shareddb.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("renja handler running without database-backed workflow: %v", err)
		return &Handler{storage: "unavailable"}
	}

	repo := renjarepo.NewRenjaGormRepository(db)
	tx := renjarepo.NewGormTxManager(db)
	svc := usecase.NewService(tx, repo)

	return &Handler{service: svc, storage: "mysql"}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	renja := v1.Group("/renja")
	renja.GET("/overview", h.GetRenjaOverview)
	renja.POST("/:id/submit", h.Submit)
	renja.POST("/:id/approve", h.Approve)
	renja.POST("/:id/reject", h.Reject)
	renja.GET("/export/indikator-csv", h.ExportIndikatorKinerjaCSV)
}
// ExportIndikatorKinerjaCSV handles CSV export for Indikator Kinerja.
func (h *Handler) ExportIndikatorKinerjaCSV(c *gin.Context) {
	rkIDStr := c.Query("rencana_kerja_id")
	unitIDStr := c.Query("unit_pengusul_id")
	if rkIDStr == "" || unitIDStr == "" {
		response.Error(c, http.StatusBadRequest, "rencana_kerja_id and unit_pengusul_id are required")
		return
	}
	rkID, err1 := strconv.ParseUint(rkIDStr, 10, 64)
	unitID, err2 := strconv.ParseUint(unitIDStr, 10, 64)
	if err1 != nil || err2 != nil || rkID == 0 || unitID == 0 {
		response.Error(c, http.StatusBadRequest, "invalid rencana_kerja_id or unit_pengusul_id")
		return
	}

	csvBytes, err := h.service.GenerateIndikatorCSV(c.Request.Context(), uint(rkID), uint(unitID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=indikator_kinerja.csv")
	c.Header("Content-Type", "text/csv")
	c.Data(http.StatusOK, "text/csv", csvBytes)
}

func (h *Handler) GetRenjaOverview(c *gin.Context) {
	response.Success(c, gin.H{
		"module":  "Renja",
		"scope":   "Perencanaan kerja tahunan RSUD",
		"status":  http.StatusText(http.StatusOK),
		"storage": h.storage,
	})
}

func (h *Handler) Submit(c *gin.Context) {
	renjaID, ok := parseID(c)
	if !ok {
		return
	}

	var req actionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.ActorID <= 0 {
		response.Error(c, http.StatusBadRequest, "actor_id is required")
		return
	}

	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "renja service unavailable")
		return
	}

	if err := h.service.Submit(c.Request.Context(), renjaID, req.ActorID); err != nil {
		h.handleUsecaseError(c, err)
		return
	}

	response.Success(c, gin.H{"id": renjaID, "action": "submit"})
}

func (h *Handler) Approve(c *gin.Context) {
	renjaID, ok := parseID(c)
	if !ok {
		return
	}

	var req actionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.ActorID <= 0 {
		response.Error(c, http.StatusBadRequest, "actor_id is required")
		return
	}

	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "renja service unavailable")
		return
	}

	if err := h.service.Approve(c.Request.Context(), renjaID, req.ActorID); err != nil {
		h.handleUsecaseError(c, err)
		return
	}

	response.Success(c, gin.H{"id": renjaID, "action": "approve"})
}

func (h *Handler) Reject(c *gin.Context) {
	renjaID, ok := parseID(c)
	if !ok {
		return
	}

	var req actionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.ActorID <= 0 {
		response.Error(c, http.StatusBadRequest, "actor_id is required")
		return
	}

	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "renja service unavailable")
		return
	}

	if err := h.service.Reject(c.Request.Context(), renjaID, req.ActorID, req.Reason); err != nil {
		h.handleUsecaseError(c, err)
		return
	}

	response.Success(c, gin.H{"id": renjaID, "action": "reject"})
}

func (h *Handler) handleUsecaseError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "renja not found")
		return
	}
	if errors.Is(err, domain.ErrInvalidTransition) || errors.Is(err, domain.ErrRejectionReasonMissing) {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Error(c, http.StatusInternalServerError, "failed to process renja action")
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}

	return id, true
}
