package bootstrap

import (
	dokumenpdfhttp "e-plan-ai/internal/modules/dokumen_pdf/delivery/http"

	"e-plan-ai/internal/config"
	authhttp "e-plan-ai/internal/modules/auth/delivery/http"
	clienthttp "e-plan-ai/internal/modules/client/delivery/http"
	crudhttp "e-plan-ai/internal/modules/crud/delivery/http"
	performancehttp "e-plan-ai/internal/modules/performance/delivery/http"
	planninghttp "e-plan-ai/internal/modules/planning/delivery/http"
	planninggormhttp "e-plan-ai/internal/modules/planninggorm/delivery/http"
	renjahttp "e-plan-ai/internal/modules/renja/delivery/http"
	uihttp "e-plan-ai/internal/modules/ui/delivery/http"
	unitpelaksanahttp "e-plan-ai/internal/modules/unitpelaksana/delivery/http"
	unitpengusulhttp "e-plan-ai/internal/modules/unitpengusul/delivery/http"
	visihttp "e-plan-ai/internal/modules/visi/delivery/http"
	"e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	devActor := resolveDevelopmentActor(cfg)

	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery(), middleware.CORS())
	r.Static("/assets", "web/assets")
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/ui/dokumen_pdf")
	})
	uihttp.NewHandler().RegisterRoutes(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":                "e-plan-ai",
			"status":                 "ok",
			"auth_enabled":           cfg.AuthEnabled,
			"dev_auth_actor_enabled": devActor.Enabled(),
		})
	})

	r.GET("/ready", func(c *gin.Context) {
		if err := database.PingMySQL(cfg); err != nil {
			c.JSON(503, gin.H{
				"service":                "e-plan-ai",
				"status":                 "degraded",
				"auth_enabled":           cfg.AuthEnabled,
				"dev_auth_actor_enabled": devActor.Enabled(),
				"database":               "unavailable",
				"error":                  "database connection unavailable",
			})
			return
		}

		c.JSON(200, gin.H{
			"service":                "e-plan-ai",
			"status":                 "ready",
			"auth_enabled":           cfg.AuthEnabled,
			"dev_auth_actor_enabled": devActor.Enabled(),
			"database":               "ok",
		})
	})

	authGroup := r.Group("/api/v1/auth")
	authhttp.NewHandler(cfg).RegisterRoutes(authGroup)

	v1 := r.Group("/api/v1")
	v1.Use(
		middleware.Auth(cfg.AuthEnabled, cfg.AuthToken),
		middleware.DevelopmentActor(cfg.AuthEnabled, devActor),
		middleware.OperatorReadOnly(cfg.AuthEnabled),
	)
	crudhttp.NewHandler(cfg).RegisterRoutes(v1)
	clienthttp.NewHandler(cfg).RegisterRoutes(v1)
	planninghttp.NewHandler().RegisterRoutes(v1)
	renjahttp.NewHandler(cfg).RegisterRoutes(v1)
	unitpelaksanahttp.NewHandler(cfg).RegisterRoutes(v1)
	unitpengusulhttp.NewHandler(cfg).RegisterRoutes(v1)
	visihttp.NewHandler(cfg).RegisterRoutes(v1)
	planninggormhttp.NewHandler(cfg).RegisterRoutes(v1)
	performancehttp.NewHandler(cfg).RegisterRoutes(v1)

	db, _ := database.NewGormMySQL(cfg)
	dokumenPDFHandler := dokumenpdfhttp.NewHandler(db)

	v1.GET("/dokumen_pdf/latest", dokumenPDFHandler.GetLatestDokumenPDF)
	dokumenPDFHandler.RegisterRoutes(v1)

	r.GET("/ui/dokumen_pdf", dokumenpdfhttp.DokumenPDFPageHandler(cfg))

	return r


}
