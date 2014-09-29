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

var rp *redis.Pool
var mg mailgun.Mailgun

func init() {
	rp = ttt.NewRedisPool()
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
	rc := rp.Get()
	defer rc.Close()
	now := time.Now().Round(time.Minute)
	emails, _ := redis.Strings(rc.Do(
		"ZRANGEBYSCORE",
		ttt.Key("deliveries"),
		0,
		now.Unix(),
	))
	for _, email := range emails {
		var err error
		var sh string

		for t := 10; t > 0; t-- {
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
				sh = ""
				continue
			}

			// Get story details
			var story ttt.Story
			v, err := redis.Values(rc.Do("HGETALL", sh))
			redis.ScanStruct(v, &story)
			if err != nil {
				log.Printf("  Error: %v", err)
				sh = ""
				continue
			}
			log.Printf("  Got %v", sh)

			// Check if story is from the same person
			if email == story.Email {
				log.Printf("  It's the same person!")
				sh = ""
				continue
			}

			// Check if story has already been delivered
			_, err = redis.Int64(rc.Do(
				"ZSCORE",
				ttt.Key("delivered", email),
				sh,
			))
			if err != redis.ErrNil {
				log.Printf("  But it was already delivered to %v", email)
				sh = ""
				continue
			}

			break
		}

		// Try again next batch
		if sh == "" {
			log.Printf("Skipping %v", email)
			continue
		}

		d := Delivery{
			Recipient: email,
			StoryHash: sh,
		}
		select {
		case deliveries <- d:
		case <-done:
			return
		}
	}
}

func sendStory(d Delivery) error {
	rc := rp.Get()
	defer rc.Close()
	var story ttt.Story
	v, err := redis.Values(rc.Do("HGETALL", d.StoryHash))
	redis.ScanStruct(v, &story)
	if err != nil {
		log.Printf("Error getting story details: %v", err)
		return err
	}
	m := mg.NewMessage(
		"Turtle <turtle@telltheturtle.com>",
		"The turtle has a story for you",
		fmt.Sprintf(`
Topic: %s

%s

Tell the turtle
http://telltheturtle.com
`, story.Topic, story.Body),
		d.Recipient,
	)
	mg.Send(m)
	rc.Do(
		"ZREM",
		ttt.Key("deliveries"),
		d.Recipient,
	)
	rc.Do(
		"ZADD",
		ttt.Key("delivered", d.Recipient),
		time.Now().Unix(),
		d.StoryHash,
	)
	log.Printf("Sent: %v", d)
	return nil
}

func sendDeliveries(
	done <-chan os.Signal,
	deliveries chan Delivery,
) {
	var err error
	for {
		select {
		case d := <-deliveries:
			err = sendStory(d)
			if err != nil {
				continue
			}
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
		case <-time.After(10 * time.Second):
		}
	}
}
