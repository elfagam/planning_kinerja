package http

import "github.com/gin-gonic/gin"

type page struct {
	Route string
	File  string
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/ui", func(c *gin.Context) {
		c.File("web/templates/index.html")
	})

	registerShortRoutes(r)

	for _, p := range pages() {
		page := p
		r.GET("/ui/"+page.Route, func(c *gin.Context) {
			if page.Route == "kontrol-pagu" {
				c.Redirect(302, "/spa/pagu-control")
				return
			}
			c.File("web/templates/" + page.File)
		})
	}
}

func pages() []page {
	return []page{
		{Route: "login", File: "login.html"},
		{Route: "visi", File: "visi.html"},
		{Route: "misi", File: "misi.html"},
		{Route: "tujuan", File: "tujuan.html"},
		{Route: "indikator-tujuan", File: "indikator-tujuan.html"},
		{Route: "sasaran", File: "sasaran.html"},
		{Route: "indikator-sasaran", File: "indikator-sasaran.html"},
		{Route: "program", File: "program.html"},
		{Route: "indikator-program", File: "indikator-program.html"},
		{Route: "kegiatan", File: "kegiatan.html"},
		{Route: "indikator-kegiatan", File: "indikator-kegiatan.html"},
		{Route: "sub-kegiatan", File: "sub-kegiatan.html"},
		{Route: "pagu-sub-kegiatan", File: "pagu-sub-kegiatan.html"},
		{Route: "unit-pengusul", File: "unit-pengusul.html"},
		{Route: "unit_pengusul", File: "unit-pengusul.html"},
		{Route: "indikator-sub-kegiatan", File: "indikator-sub-kegiatan.html"},
		{Route: "rencana-kerja", File: "rencana-kerja.html"},
		{Route: "rencana-kerja-spa", File: "rencana-kerja-spa.html"},
		{Route: "renja", File: "renja.html"},
		{Route: "indikator-kinerja", File: "indikator-kinerja.html"},
		{Route: "target-realisasi", File: "target-realisasi.html"},
		{Route: "target-evaluasi", File: "target-realisasi.html"},
		{Route: "informasi", File: "informasi.html"},
		{Route: "manajemen-user", File: "manajemen-user.html"},
		{Route: "clients", File: "clients.html"},
		{Route: "dokumen_pdf", File: "dokumen_pdf.html"},
		{Route: "dashboard", File: "dashboard.html"},
		{Route: "kontrol-pagu", File: "kontrol_pagu.html"},
		{Route: "qna", File: "qna.html"},
	}
}

func registerShortRoutes(r *gin.Engine) {
	shortToUI := map[string]string{
		"/dashboard":       "/ui/dashboard",
		"/rencana-kerja":   "/ui/rencana-kerja",
		"/target-evaluasi": "/ui/target-evaluasi",
		"/informasi":       "/ui/informasi",
		"/qna":             "/ui/qna",
	}

	for shortPath, uiPath := range shortToUI {
		path := shortPath
		target := uiPath
		r.GET(path, func(c *gin.Context) {
			c.Redirect(302, target)
		})
	}
}
