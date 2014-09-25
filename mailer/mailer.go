package mailer

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/mailgun/mailgun-go"
	"github.com/marksteve/telltheturtle"
)

var rc redis.Conn
var mg mailgun.Mailgun

func init() {
	var err error
	rc, err = redis.Dial("tcp", "redis:6379")
	if err != nil {
		panic(err)
	}
	mg = mailgun.NewMailgun(
		"telltheturtle.com",
		os.Getenv("MAILGUN_API_KEY"),
		"",
	)
}

type Delivery struct {
	Recipient string
	StoryHash string
}

func getPendingDeliveries(
	done <-chan os.Signal,
	deliveries chan Delivery,
) {
	now := time.Now().Round(time.Minute)
	emails, _ := redis.Strings(rc.Do(
		"ZRANGEBYSCORE",
		ttt.Key("deliveries"),
		0,
		now.Unix(),
	))
	for _, email := range emails {
		sh, _ := redis.String(rc.Do(
			"SRANDMEMBER",
			ttt.Key("stories"),
		))
		delivery := Delivery{
			Recipient: email,
			StoryHash: sh,
		}
		select {
		case deliveries <- delivery:
		case <-done:
			return
		}
	}
}

type Story struct {
	Topic string `redis:"topic"`
	Body  string `redis:"body"`
}

func sendDeliveries(
	done <-chan os.Signal,
	deliveries chan Delivery,
) {
	for {
		select {
		case delivery := <-deliveries:
			v, _ := redis.Values(rc.Do("HGETALL", delivery.StoryHash))
			var story Story
			redis.ScanStruct(v, &story)
			m := mg.NewMessage(
				"Turtle <turtle@telltheturtle.com>",
				"The turtle has a story for you",
				fmt.Sprintf(`
Topic: %s

%s
`, story.Topic, story.Body),
				delivery.Recipient,
			)
			mg.Send(m)
			rc.Do(
				"ZREM",
				ttt.Key("deliveries"),
				delivery.Recipient,
			)
			log.Printf("Sent: %v", delivery)
		case <-done:
			return
		}
	}
}

func Run(done chan os.Signal) {
	log.Println("Started mailer")
	deliveries := make(chan Delivery)
	go sendDeliveries(done, deliveries)
	for {
		go getPendingDeliveries(done, deliveries)
		select {
		case <-done:
			log.Println("Bye.")
			return
		case <-time.After(time.Minute):
		}
	}
}
