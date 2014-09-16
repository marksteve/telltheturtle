package web

import (
	"github.com/zenazn/goji"
)

func Serve() {
	goji.Get("/", Index)
	goji.Serve()
}
