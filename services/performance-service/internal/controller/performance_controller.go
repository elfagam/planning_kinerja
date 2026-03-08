package controller

import (
	"net/http"

	"eplan/services/performance-service/internal/service"

	"github.com/gin-gonic/gin"
)

type PerformanceController struct {
	svc service.PerformanceService
}

func NewPerformanceController(svc service.PerformanceService) PerformanceController {
	return PerformanceController{svc: svc}
}

func (c PerformanceController) Register(r gin.IRouter) {
	r.GET("/v1/target-realisasi", c.list)
}

func (c PerformanceController) list(ctx *gin.Context) {
	items, err := c.svc.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load performance"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}
