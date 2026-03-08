package controller

import (
	"net/http"

	"eplan/services/renja-service/internal/service"

	"github.com/gin-gonic/gin"
)

type RenjaController struct {
	svc service.RenjaService
}

func NewRenjaController(svc service.RenjaService) RenjaController {
	return RenjaController{svc: svc}
}

func (c RenjaController) Register(r gin.IRouter) {
	r.GET("/v1/renja", c.list)
}

func (c RenjaController) list(ctx *gin.Context) {
	items, err := c.svc.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load renja"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}
