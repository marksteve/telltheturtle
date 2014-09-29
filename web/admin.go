package web

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

func Admin(c web.C, w http.ResponseWriter, r *http.Request) {
	var msg string
	rc := rp.Get()

	defer func() {
		t.ExecuteTemplate(w, "Admin", map[string]interface{}{
			"Message": msg,
		})

		rc.Close()
	}()

	if r.Method == "POST" {
	}
}
