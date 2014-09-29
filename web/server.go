package web

import (
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/marksteve/telltheturtle"
	"github.com/zenazn/goji"
)

var t *template.Template
var rp *redis.Pool

func init() {
	rand.Seed(time.Now().UnixNano())
	t = template.Must(template.ParseGlob("web/templates/*.html"))
	rp = ttt.NewRedisPool()
}

func Serve() {
	goji.Get("/", Index)
	goji.Post("/", Index)

	goji.Get("/supersecret", Admin)
	goji.Post("/supersecret", Admin)

	goji.Get("/remove_story/:sh", RemoveStory)

	goji.Get("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("web/static")),
	))
	goji.Serve()
}
