package messaging

type Event interface {
	Key() string
}
