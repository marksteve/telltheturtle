package ttt

type Story struct {
	Topic string `redis:"topic"`
	Body  string `redis:"body"`
	Email string `redis:"email"`
}
