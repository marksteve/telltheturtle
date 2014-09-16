package web

import (
	"fmt"
	"net/http"

	"github.com/zenazn/goji/web"
)

func Index(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world")
}
