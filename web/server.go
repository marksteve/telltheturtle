package web

import (
	"net/http"

	"github.com/zenazn/goji"
)

func Serve() {
	goji.Get("/", Index)
	goji.Post("/", Index)

	goji.Get("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("web/static")),
	))
	goji.Serve()
}
