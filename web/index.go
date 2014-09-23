package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/zenazn/goji/web"
)

var t *template.Template
var rc redis.Conn

func init() {
	var err error
	t = template.Must(template.ParseGlob("templates/*.html"))
	rc, err = redis.Dial("tcp", "redis:6379")
	if err != nil {
		panic(err)
	}
}

func Index(c web.C, w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == "POST" {
		if err = r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		storyId := Key("kmsp", "story", GenID())

		// Save story
		story := []interface{}{storyId}
		for n, v := range r.PostForm {
			if len(v[0]) < 1 {
				http.Error(
					w,
					fmt.Sprintf("%s cannot be empty", strings.Title(n)),
					http.StatusBadRequest,
				)
				return
			}
			story = append(story, n, v[0])
		}
		_, err = rc.Do("HMSET", story...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Add to story pool
		_, err = rc.Do("SADD", Key("kmsp", "stories"), storyId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	t.ExecuteTemplate(w, "Index", nil)
}
