package controller

import (
	"net/http"

	"eplan/services/planning-service/internal/service"

	"github.com/gin-gonic/gin"
)

type StrategicController struct {
	svc service.StrategicService
}

func NewStrategicController(svc service.StrategicService) StrategicController {
	return StrategicController{svc: svc}
}

func (c StrategicController) Register(r gin.IRouter) {
	r.GET("/v1/strategic/:type", c.list)
}

func (c StrategicController) list(ctx *gin.Context) {
	items, err := c.svc.ListByType(ctx.Param("type"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load data"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}
