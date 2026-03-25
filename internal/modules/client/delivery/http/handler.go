package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/modules/client/domain"
	clientrepo "e-plan-ai/internal/modules/client/repository"
	"e-plan-ai/internal/modules/client/usecase"
	shareddb "e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service *usecase.Service
	storage string
}

func NewHandler(cfg config.Config) *Handler {
	db, err := shareddb.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("client handler running without database-backed service: %v", err)
		return &Handler{storage: "unavailable"}
	}

	repo := clientrepo.NewGormRepository(db)
	tx := clientrepo.NewGormTxManager(db)
	svc := usecase.NewService(tx, repo)
	return &Handler{service: svc, storage: "mysql"}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	g := v1.Group("/clients")
	g.GET("", h.List)
	g.GET("/audit-logs", h.AuditLogs)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)

	g.POST("/:id/submit", h.Submit)
	g.POST("/:id/unsubmit", h.Unsubmit)
	g.POST("/:id/reject", h.Reject)
	g.POST("/:id/re-evaluate", h.ReEvaluate)
	g.POST("/:id/approve", h.Approve)
	g.GET("/:id/status-history", h.StatusHistory)
}

func (h *Handler) AuditLogs(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "client service unavailable")
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		return
	}

	filter := usecase.AuditListFilter{
		Action: strings.TrimSpace(c.Query("action")),
		Page:   parsePositiveInt(c.Query("page"), 1),
		Limit:  parsePositiveInt(c.Query("limit"), 10),
	}

	if userRaw := strings.TrimSpace(c.Query("user_id")); userRaw != "" {
		userID, err := strconv.ParseUint(userRaw, 10, 64)
		if err != nil || userID == 0 {
			response.Error(c, http.StatusBadRequest, "user_id harus berupa angka > 0")
			return
		}
		v := uint64(userID)
		filter.UserID = &v
	}

	if resourceRaw := strings.TrimSpace(c.Query("resource_id")); resourceRaw != "" {
		resourceID, err := strconv.ParseUint(resourceRaw, 10, 64)
		if err != nil || resourceID == 0 {
			response.Error(c, http.StatusBadRequest, "resource_id harus berupa angka > 0")
			return
		}
		v := uint64(resourceID)
		filter.ResourceID = &v
	}

	items, total, err := h.service.ListAuditLogs(c.Request.Context(), actor, filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"items": items,
		"meta": gin.H{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
		},
	})
}

func (h *Handler) List(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "client service unavailable")
		return
	}

	filter := usecase.ListFilter{
		Q:      strings.TrimSpace(c.Query("q")),
		Status: strings.TrimSpace(c.Query("status")),
		Page:   parsePositiveInt(c.Query("page"), 1),
		Limit:  parsePositiveInt(c.Query("limit"), 10),
	}
	if unitRaw := strings.TrimSpace(c.Query("unit_pengusul_id")); unitRaw != "" {
		unitID, err := strconv.ParseUint(unitRaw, 10, 64)
		if err != nil || unitID == 0 {
			response.Error(c, http.StatusBadRequest, "unit_pengusul_id harus berupa angka > 0")
			return
		}
		v := uint64(unitID)
		filter.UnitPengusulID = &v
	}

	actor, ok := actorFromContext(c)
	if !ok {
		return
	}

	items, total, err := h.service.List(c.Request.Context(), actor, filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"items": items,
		"meta": gin.H{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
		},
		"storage": h.storage,
	})
}

func (h *Handler) Create(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "client service unavailable")
		return
	}

	var req upsertClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, bindErrorMessage(err, "invalid payload"))
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		return
	}

	created, err := h.service.Create(c.Request.Context(), actor, domain.Client{
		Kode:           req.Kode,
		Nama:           req.Nama,
		UnitPengusulID: req.UnitPengusulID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, created)
}

func (h *Handler) Get(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "client service unavailable")
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		return
	}

	item, err := h.service.Get(c.Request.Context(), actor, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) Update(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "client service unavailable")
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	var req upsertClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, bindErrorMessage(err, "invalid payload"))
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		return
	}

	updated, err := h.service.Update(c.Request.Context(), id, actor, domain.Client{
		Kode:           req.Kode,
		Nama:           req.Nama,
		UnitPengusulID: req.UnitPengusulID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "client service unavailable")
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id, actor); err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true})
}

func (h *Handler) Submit(c *gin.Context) {
	h.handleTransition(c, func(id uint64, actor usecase.Actor, req transitionRequest) error {
		return h.service.Submit(c.Request.Context(), id, actor, req.Note)
	}, "submit")
}

func (h *Handler) Unsubmit(c *gin.Context) {
	h.handleTransition(c, func(id uint64, actor usecase.Actor, req transitionRequest) error {
		return h.service.Unsubmit(c.Request.Context(), id, actor, req.Reason)
	}, "unsubmit")
}

func (h *Handler) Reject(c *gin.Context) {
	h.handleTransition(c, func(id uint64, actor usecase.Actor, req transitionRequest) error {
		return h.service.Reject(c.Request.Context(), id, actor, req.Reason)
	}, "reject")
}

func (h *Handler) ReEvaluate(c *gin.Context) {
	h.handleTransition(c, func(id uint64, actor usecase.Actor, req transitionRequest) error {
		return h.service.ReEvaluate(c.Request.Context(), id, actor, req.Reason)
	}, "re-evaluate")
}

func (h *Handler) Approve(c *gin.Context) {
	h.handleTransition(c, func(id uint64, actor usecase.Actor, req transitionRequest) error {
		return h.service.Approve(c.Request.Context(), id, actor, req.Note)
	}, "approve")
}

func (h *Handler) StatusHistory(c *gin.Context) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "client service unavailable")
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	items, err := h.service.StatusHistory(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

func (h *Handler) handleTransition(c *gin.Context, fn func(id uint64, actor usecase.Actor, req transitionRequest) error, action string) {
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "client service unavailable")
		return
	}

	id, ok := parseID(c)
	if !ok {
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		return
	}

	var req transitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, bindErrorMessage(err, "invalid payload"))
		return
	}
	if err := fn(id, actor, req); err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "action": action})
}

func actorFromContext(c *gin.Context) (usecase.Actor, bool) {
	rawID, ok := c.Get("auth.user_id")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "missing auth user id")
		return usecase.Actor{}, false
	}

	actorID, ok := toUint64(rawID)
	if !ok || actorID == 0 {
		response.Error(c, http.StatusUnauthorized, "invalid auth user id")
		return usecase.Actor{}, false
	}

	rawRole, ok := c.Get("auth.role")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "missing auth role")
		return usecase.Actor{}, false
	}
	role := strings.TrimSpace(fmt.Sprintf("%v", rawRole))
	if role == "" {
		response.Error(c, http.StatusUnauthorized, "invalid auth role")
		return usecase.Actor{}, false
	}

	rawName, _ := c.Get("auth.full_name")
	name := strings.TrimSpace(fmt.Sprintf("%v", rawName))

	return usecase.Actor{
		ID:        actorID,
		Role:      role,
		Name:      name,
		IPAddress: strings.TrimSpace(c.ClientIP()),
		UserAgent: strings.TrimSpace(c.Request.UserAgent()),
	}, true
}

func toUint64(v any) (uint64, bool) {
	switch n := v.(type) {
	case uint64:
		return n, true
	case uint32:
		return uint64(n), true
	case uint:
		return uint64(n), true
	case int64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case float64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(n), 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrClientNotFound):
		response.Error(c, http.StatusNotFound, "client not found")
	case errors.Is(err, domain.ErrInvalidStatus):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrReasonRequired):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrValidation):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidTransition):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrForbiddenOperation):
		response.Error(c, http.StatusForbidden, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("failed to process client request: %v", err))
	}
}

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func parsePositiveInt(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func bindErrorMessage(err error, fallback string) string {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) || len(validationErrs) == 0 {
		return fallback
	}

	fe := validationErrs[0]
	field := validationFieldName(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s minimum length is %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s maximum length is %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	default:
		return fallback
	}
}

func validationFieldName(raw string) string {
	switch raw {
	case "Kode":
		return "kode"
	case "Nama":
		return "nama"
	case "UnitPengusulID":
		return "unit_pengusul_id"
	case "Reason":
		return "reason"
	case "Note":
		return "note"
	default:
		return strings.ToLower(raw)
	}
}
