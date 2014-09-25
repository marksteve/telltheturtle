package web

import (
	"errors"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/zenazn/goji/web"
)

var t *template.Template
var rc redis.Conn
var topics = []string{
	"A time where you truly felt alive",
	"A moment when you thought you were going to die",
	"A deep realization you had",
	"Your best college/high school memory",
	"Your first heartbreak",
	"Your last heartbreak",
	"Your deepest regret",
	"Your dependence on a substance to cope with life",
	"Your lowest point",
	"Your highest point",
	"How you realized your calling in life",
	"A twisted fairy tale",
	"A dark topic in the form of a children's story",
	"What if a mythological hero was transported to the present?",
}

func getTopic() string {
	return topics[rand.Intn(len(topics))]
}

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
	var msg string

	defer func() {
		var topic string
		if err == nil {
			topic = getTopic()
		} else {
			topic = r.PostFormValue("topic")
		}

		t.ExecuteTemplate(w, "Index", map[string]string{
			"Topic":   topic,
			"Message": msg,
		})
	}()

	if r.Method == "POST" {
		if err = r.ParseForm(); err != nil {
			msg = err.Error()
			return
		}

		storyId := Key("kmsp", "story", GenID())

		// Save story
		story := []interface{}{storyId}
		for n, v := range r.PostForm {
			switch n {
			case "body":
				if len(v[0]) < 150 {
					err = errors.New("Too short. At least give me a tweet long!")
				}
			case "email":
				if len(v[0]) < 1 {
					err = errors.New("I need your email so I can send you stuff life friends do.")
				}
			}
			if err != nil {
				msg = err.Error()
				return
			}
			story = append(story, n, v[0])
		}
		_, err = rc.Do("HMSET", story...)
		if err != nil {
			msg = err.Error()
			return
		}

		// Add to story pool
		_, err = rc.Do("SADD", Key("kmsp", "stories"), storyId)
		if err != nil {
			msg = err.Error()
			return
		}
		msg = "Got it!"
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
