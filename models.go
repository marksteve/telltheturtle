package ttt

type Story struct {
	Hash  string
	Topic string `redis:"topic"`
	Body  string `redis:"body"`
	Email string `redis:"email"`
}
