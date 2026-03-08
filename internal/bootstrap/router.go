package bootstrap

import (
	"e-plan-ai/internal/config"
	crudhttp "e-plan-ai/internal/modules/crud/delivery/http"
	performancehttp "e-plan-ai/internal/modules/performance/delivery/http"
	planninghttp "e-plan-ai/internal/modules/planning/delivery/http"
	planninggormhttp "e-plan-ai/internal/modules/planninggorm/delivery/http"
	renjahttp "e-plan-ai/internal/modules/renja/delivery/http"
	uihttp "e-plan-ai/internal/modules/ui/delivery/http"
	unitpelaksanahttp "e-plan-ai/internal/modules/unitpelaksana/delivery/http"
	unitpengusulhttp "e-plan-ai/internal/modules/unitpengusul/delivery/http"
	visihttp "e-plan-ai/internal/modules/visi/delivery/http"
	"e-plan-ai/internal/shared/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery())
	r.Static("/assets", "web/assets")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/ui")
	})
	uihttp.NewHandler().RegisterRoutes(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "e-plan-ai", "status": "ok"})
	})

	v1 := r.Group("/api/v1")
	v1.Use(middleware.Auth(cfg.AuthEnabled, cfg.AuthToken))
	crudhttp.NewHandler(cfg).RegisterRoutes(v1)
	planninghttp.NewHandler().RegisterRoutes(v1)
	renjahttp.NewHandler(cfg).RegisterRoutes(v1)
	unitpelaksanahttp.NewHandler(cfg).RegisterRoutes(v1)
	unitpengusulhttp.NewHandler(cfg).RegisterRoutes(v1)
	visihttp.NewHandler(cfg).RegisterRoutes(v1)
	planninggormhttp.NewHandler(cfg).RegisterRoutes(v1)
	performancehttp.NewHandler(cfg).RegisterRoutes(v1)

	return r
}
