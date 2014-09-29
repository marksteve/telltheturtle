package web

import (
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/marksteve/telltheturtle"
	"github.com/zenazn/goji/web"
)

func Admin(c web.C, w http.ResponseWriter, r *http.Request) {
	var err error
	var msg string
	rc := rp.Get()

	shs, err := redis.Strings(rc.Do(
		"SMEMBERS",
		ttt.Key("stories"),
	))
	var stories []ttt.Story
	for _, sh := range shs {
		var story ttt.Story
		v, err := redis.Values(rc.Do("HGETALL", sh))
		if err != nil {
			continue
		}
		redis.ScanStruct(v, &story)
		story.Hash = sh
		stories = append(stories, story)
	}

	defer func() {
		if err != nil {
			msg = err.Error()
		}

		t.ExecuteTemplate(w, "Admin", map[string]interface{}{
			"Message": msg,
			"Stories": stories,
		})

		rc.Close()
	}()

	if r.Method == "POST" {
	}
}

func RemoveStory(c web.C, w http.ResponseWriter, r *http.Request) {
	rc := rp.Get()
	sh := c.URLParams["sh"]
	rc.Do("SREM", ttt.Key("stories"), sh)
	rc.Do("SADD", ttt.Key("removed_stories"), sh)
	http.Redirect(w, r, "/supersecret", http.StatusTemporaryRedirect)
}
