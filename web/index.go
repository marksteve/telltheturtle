package web

import (
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/marksteve/telltheturtle"
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
	rand.Seed(time.Now().UnixNano())
	var err error
	t = template.Must(template.ParseGlob("web/templates/*.html"))
	rc, err = redis.Dial("tcp", "redis:6379")
	if err != nil {
		panic(err)
	}
}

func Index(c web.C, w http.ResponseWriter, r *http.Request) {
	var err error
	var msg string

	defer func() {
		var topic, body, email string
		if err == nil {
			topic = getTopic()
			body = ""
			email = ""
		} else {
			topic = r.PostFormValue("topic")
			body = r.PostFormValue("body")
			email = r.PostFormValue("email")
		}

		sc, _ := redis.Int(rc.Do("SCARD", ttt.Key("stories")))

		t.ExecuteTemplate(w, "Index", map[string]interface{}{
			"Message":      msg,
			"Topic":        topic,
			"Body":         body,
			"Email":        email,
			"StoriesCount": sc,
		})
	}()

	if r.Method == "POST" {
		if err = r.ParseForm(); err != nil {
			msg = err.Error()
			return
		}

		storyId := ttt.Key("story", ttt.GenID())

		// Save story
		story := []interface{}{storyId}
		for n, v := range r.PostForm {
			_v := v[0]
			switch n {
			case "body":
				_v = strings.Trim(_v, " ")
				if len(_v) < 150 {
					err = errors.New("Too short. At least give me a tweet long!")
				}
			case "email":
				_v = strings.Trim(strings.ToLower(_v), " ")
				if len(_v) < 1 {
					err = errors.New("I need your email so I can send you stuff like friends do.")
				}
			}
			if err != nil {
				msg = err.Error()
				return
			}
			story = append(story, n, _v)
		}
		_, err = rc.Do("HMSET", story...)
		if err != nil {
			msg = err.Error()
			return
		}

		// Add to story pool
		_, err = rc.Do("SADD", ttt.Key("stories"), storyId)
		if err != nil {
			msg = err.Error()
			return
		}

		// Set delivery
		var delivery time.Time
		email := r.PostFormValue("email")
		ret, err := redis.Int64(rc.Do(
			"ZSCORE",
			ttt.Key("deliveries"),
			email,
		))
		if err == redis.ErrNil {
			delivery = time.Now().Add(
				time.Duration(4) * time.Hour,
			).Add(
				time.Duration(rand.Intn(16)) * time.Hour,
			)
			_, err = rc.Do(
				"ZADD",
				ttt.Key("deliveries"),
				delivery.Unix(),
				email,
			)
			if err != nil {
				msg = err.Error()
				return
			}
		} else {
			delivery = time.Unix(ret, 0)
		}

		msg = fmt.Sprintf(
			"Cool story bro. I'll be back in %d hours.",
			int(delivery.Sub(time.Now()).Hours()),
		)
	}
}
