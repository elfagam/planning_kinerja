package http

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/modules/crud/domain"
	"e-plan-ai/internal/modules/crud/repository"
	"e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store repository.Store
}

func NewHandler(cfg config.Config) *Handler {
	var store repository.Store
	mysqlStore, err := repository.NewMySQLStore(cfg.MySQLDSN)
	if err != nil {
		if database.IsConnectionError(err) {
			log.Printf("crud handler using in-memory store due to database connection failure: %v", err)
		} else {
			log.Printf("crud handler using in-memory store due to mysql store init failure: %v", err)
		}
		store = repository.NewMemoryStore(resourceKeys())
	} else {
		store = mysqlStore
	}
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	for _, key := range resourceKeys() {
		group := v1.Group("/" + key)
		group.GET("", h.list(key))
		group.POST("", h.create(key))
		group.GET("/:id", h.get(key))
		group.PUT("/:id", h.update(key))
		group.DELETE("/:id", h.delete(key))
	}
}

func (h *Handler) list(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := parseFilter(c)
		rows, total, err := h.store.List(resource, filter)
		if err != nil {
			if database.IsConnectionError(err) {
				response.Error(c, http.StatusServiceUnavailable, "database connection unavailable")
				return
			}
			response.Error(c, http.StatusInternalServerError, "failed to load data")
			return
		}

		response.Success(c, gin.H{
			"resource": resource,
			"items":    rows,
			"meta": gin.H{
				"page":  filter.Page,
				"limit": filter.Limit,
				"total": total,
			},
		})
	}
}

func (h *Handler) create(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req domain.Payload
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "invalid payload")
			return
		}
		if req.Name == "" {
			response.Error(c, http.StatusBadRequest, "name is required")
			return
		}

		item, err := h.store.Create(resource, req)
		if err != nil {
			if database.IsConnectionError(err) {
				response.Error(c, http.StatusServiceUnavailable, "database connection unavailable")
				return
			}
			response.Error(c, http.StatusInternalServerError, "failed to create data")
			return
		}
		response.Success(c, item)
	}
}

func (h *Handler) get(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid id")
			return
		}

		item, err := h.store.Get(resource, id)
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "data not found")
			return
		}
		if err != nil {
			if database.IsConnectionError(err) {
				response.Error(c, http.StatusServiceUnavailable, "database connection unavailable")
				return
			}
			response.Error(c, http.StatusInternalServerError, "failed to get data")
			return
		}
		response.Success(c, item)
	}
}

func (h *Handler) update(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid id")
			return
		}

		var req domain.Payload
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "invalid payload")
			return
		}
		if req.Name == "" {
			response.Error(c, http.StatusBadRequest, "name is required")
			return
		}

		item, err := h.store.Update(resource, id, req)
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "data not found")
			return
		}
		if err != nil {
			if database.IsConnectionError(err) {
				response.Error(c, http.StatusServiceUnavailable, "database connection unavailable")
				return
			}
			response.Error(c, http.StatusInternalServerError, "failed to update data")
			return
		}

		response.Success(c, item)
	}
}

func (h *Handler) delete(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid id")
			return
		}

		err = h.store.Delete(resource, id)
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "data not found")
			return
		}
		if err != nil {
			if database.IsConnectionError(err) {
				response.Error(c, http.StatusServiceUnavailable, "database connection unavailable")
				return
			}
			response.Error(c, http.StatusInternalServerError, "failed to delete data")
			return
		}
		response.Success(c, gin.H{"deleted": id})
	}
}

func parseFilter(c *gin.Context) domain.ListFilter {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return domain.ListFilter{
		Query:  c.Query("q"),
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func resourceKeys() []string {
	return []string{
		"indikator-kinerja",
		"target-realisasi",
	}
}
