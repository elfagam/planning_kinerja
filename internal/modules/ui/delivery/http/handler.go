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

	for _, p := range pages() {
		page := p
		r.GET("/ui/"+page.Route, func(c *gin.Context) {
			c.File("web/templates/" + page.File)
		})
	}
}

func pages() []page {
	return []page{
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
		{Route: "indikator-sub-kegiatan", File: "indikator-sub-kegiatan.html"},
		{Route: "rencana-kerja", File: "rencana-kerja.html"},
		{Route: "renja", File: "renja.html"},
		{Route: "indikator-kinerja", File: "indikator-kinerja.html"},
		{Route: "target-realisasi", File: "target-realisasi.html"},
		{Route: "dashboard", File: "dashboard.html"},
	}
}
