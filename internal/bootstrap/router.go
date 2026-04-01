package bootstrap

import (
	"e-plan-ai/internal/config"
	authhttp "e-plan-ai/internal/modules/auth/delivery/http"
	clienthttp "e-plan-ai/internal/modules/client/delivery/http"
	crudhttp "e-plan-ai/internal/modules/crud/delivery/http"
	dokumenpdfhttp "e-plan-ai/internal/modules/dokumen_pdf/delivery/http"
	performancehttp "e-plan-ai/internal/modules/performance/delivery/http"
	planninghttp "e-plan-ai/internal/modules/planning/delivery/http"
	planninggormhttp "e-plan-ai/internal/modules/planninggorm/delivery/http"
	qnahttp "e-plan-ai/internal/modules/qna/delivery/http"
	qnarepo "e-plan-ai/internal/modules/qna/repository"
	qnausecase "e-plan-ai/internal/modules/qna/usecase"
	renjahttp "e-plan-ai/internal/modules/renja/delivery/http"
	uihttp "e-plan-ai/internal/modules/ui/delivery/http"
	unitpelaksanahttp "e-plan-ai/internal/modules/unitpelaksana/delivery/http"
	unitpengusulhttp "e-plan-ai/internal/modules/unitpengusul/delivery/http"
	visihttp "e-plan-ai/internal/modules/visi/delivery/http"
	"e-plan-ai/internal/shared/database"
	"e-plan-ai/internal/shared/middleware"

	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config) *gin.Engine {
	devActor := resolveDevelopmentActor(cfg)

	r := gin.New()
	// Inisialisasi template HTML dari folder web/templates
	r.LoadHTMLGlob("web/templates/*")
	r.Use(middleware.Security(), gin.Logger(), middleware.Recovery(), middleware.CORS())
	
	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		log.Printf("[DATABASE] CRITICAL CONNECTION FAILURE: %v", err)
	} else {
		log.Printf("[DATABASE] Connection established successfully.")
	}

	r.Static("/assets", "web/assets")
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	// Serve SPA if exists (Vite build output)
	if _, err := os.Stat("frontend/dist"); err == nil {
		r.Static("/spa/assets", "frontend/dist/assets")
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// If it's a /spa/ route, return index.html
			if strings.HasPrefix(path, "/spa/") {
				c.File("frontend/dist/index.html")
				return
			}
			// Skip for API/Auth/Legacy UI routes
			if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/auth") || strings.HasPrefix(path, "/ui") || strings.HasPrefix(path, "/assets") {
				return 
			}
			// Fallback to legacy index or 404
		})
	}

	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/ui/dashboard")
	})
	uihttp.NewHandler().RegisterRoutes(r)

	// Tambahkan ini agar /spa/pagu-control mengarah ke template HTML manual
	r.GET("/spa/pagu-control", func(c *gin.Context) {
		c.HTML(200, "kontrol_pagu.html", gin.H{
			"title": "Kontrol Pagu - AI-Planning",
		})
	})

	// Jika ada halaman lain yang juga ingin diaktifkan lewat template:
	// r.GET("/spa/rencana-kerja", func(c *gin.Context) {
	// 	c.HTML(200, "rencana-kerja.html", nil)
	// })

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

	// Q&A Module
	qnaRepository := qnarepo.NewQnaGormRepository(db)
	qnaUsecase := qnausecase.NewQnaUsecase(qnaRepository)
	qnahttp.NewQnaHandler(qnaUsecase).RegisterRoutes(v1)

	dokumenPDFHandler := dokumenpdfhttp.NewHandler(db)

	v1.GET("/dokumen_pdf/latest", dokumenPDFHandler.GetLatestDokumenPDF)
	dokumenPDFHandler.RegisterRoutes(v1)


	return r


}
