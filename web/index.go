package web

import (
	"html/template"
	"net/http"

	"github.com/zenazn/goji/web"
)

var t *template.Template

func Index(c web.C, w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "Index", nil)
}

func init() {
	t = template.Must(template.ParseGlob("templates/*.html"))
}
