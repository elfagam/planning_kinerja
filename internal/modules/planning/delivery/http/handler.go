package http

import (
	"net/http"

	"e-plan-ai/internal/modules/planning/domain"
	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	planning := v1.Group("/planning")
	planning.GET("/hierarchy", h.GetHierarchy)
}

func (h *Handler) GetHierarchy(c *gin.Context) {
	response.Success(c, domain.HierarchyResponse{
		Modules: []domain.HierarchyModule{
			{Key: "visi", Label: "Visi"},
			{Key: "misi", Label: "Misi", ParentKey: "visi"},
			{Key: "tujuan", Label: "Tujuan", ParentKey: "misi"},
			{Key: "indikator_tujuan", Label: "Indikator Tujuan", ParentKey: "tujuan"},
			{Key: "sasaran", Label: "Sasaran", ParentKey: "tujuan"},
			{Key: "indikator_sasaran", Label: "Indikator Sasaran", ParentKey: "sasaran"},
			{Key: "program", Label: "Program", ParentKey: "sasaran"},
			{Key: "indikator_program", Label: "Indikator Program", ParentKey: "program"},
			{Key: "kegiatan", Label: "Kegiatan", ParentKey: "program"},
			{Key: "indikator_kegiatan", Label: "Indikator Kegiatan", ParentKey: "kegiatan"},
			{Key: "sub_kegiatan", Label: "Sub Kegiatan", ParentKey: "kegiatan"},
			{Key: "indikator_sub_kegiatan", Label: "Indikator Sub Kegiatan", ParentKey: "sub_kegiatan"},
		},
		Status: http.StatusText(http.StatusOK),
	})
}
