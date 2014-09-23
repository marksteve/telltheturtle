package web

import (
	"net/http"

	"github.com/zenazn/goji"
)

func Serve() {
	goji.Get("/", Index)
	goji.Get("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	))
	goji.Serve()
}
