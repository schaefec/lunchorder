package main

import (
	"context"
	"log"
	"time"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data string `json:"name"`
}

// HelloPubSub consumes a Pub/Sub message.
func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data) + ". It's me!... Again!!!")
	return nil
}

type mockContext struct {
}

func (*mockContext) Deadline() (time.Time, bool) {
	return time.Now(), false
}

func (*mockContext) Done() <-chan struct{} {
	return make(chan struct{})
}

func (*mockContext) Err() error {
	return nil
}

func (*mockContext) Value(v interface{}) interface{} {
	return nil
}

func main() {
	HelloPubSub(&mockContext{}, PubSubMessage{
		Data: "Test message",
	})
}
