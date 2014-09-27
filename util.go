package ttt

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/randbo"
	"github.com/garyburd/redigo/redis"
)

func NewRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "redis:6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func GenID() string {
	p := make([]byte, 8)
	randbo.New().Read(p)
	return fmt.Sprintf("%x", p)
}

func Key(args ...string) string {
	args = append([]string{"ttt"}, args...)
	return strings.Join(args, ":")
}
