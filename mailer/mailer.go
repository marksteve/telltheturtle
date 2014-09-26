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
		var err error
		sh := ""
		for t := 5; t > 0; t-- {
			// Some delay to avoid choking redis
			time.Sleep(100 * time.Millisecond)

			// Get random story hash
			log.Printf("Getting a random story for %v", email)
			sh, err = redis.String(rc.Do(
				"SRANDMEMBER",
				ttt.Key("stories"),
			))
			if err != nil {
				log.Printf("  Error: %v", err)
				continue
			}

			// Get story details
			var story Story
			err = getStory(sh, &story)
			if err != nil {
				log.Printf("  Error: %v", err)
				continue
			}
			log.Printf("  Got %v", sh)

			// Check if story is from the same person
			if email == story.Email {
				log.Printf("  It's the same person!")
				continue
			}

			// Check if story has already been delivered
			_, err := redis.Int64(rc.Do(
				"ZSCORE",
				ttt.Key("delivered", email),
				sh,
			))
			if err != redis.ErrNil {
				log.Printf("  But it was already delivered to %v", email)
				continue
			}

			break
		}

		// Try again next batch
		if sh == "" {
			log.Printf("  Skipping %v", email)
			continue
		}

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
	Email string `redis:"email"`
}

func getStory(sh string, story *Story) error {
	v, err := redis.Values(rc.Do("HGETALL", sh))
	redis.ScanStruct(v, story)
	return err
}

func sendDeliveries(
	done <-chan os.Signal,
	deliveries chan Delivery,
) {
	var err error
	for {
		select {
		case delivery := <-deliveries:
			var story Story
			err = getStory(delivery.StoryHash, &story)
			if err != nil {
				log.Printf("Error getting story details: %v", err)
				continue
			}
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
			rc.Do(
				"ZADD",
				ttt.Key("delivered", delivery.Recipient),
				time.Now().Unix(),
				delivery.StoryHash,
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
