package events

import (
	"sync"
	"testing"
)

func TestUnsubscribeRemovesAndClosesChannel(t *testing.T) {
	eb := NewEventBus()
	a := eb.Subscribe(RunningOrderUpdated, 1)
	b := eb.Subscribe(RunningOrderUpdated, 1)

	eb.Unsubscribe(RunningOrderUpdated, a)

	if n := eb.SubscriberCount(RunningOrderUpdated); n != 1 {
		t.Fatalf("SubscriberCount = %d, want 1", n)
	}
	if _, ok := <-a; ok {
		t.Error("unsubscribed channel is still open")
	}

	eb.Publish(Event{Type: RunningOrderUpdated, Payload: "RO1"})
	if ev := <-b; ev.Payload != "RO1" {
		t.Errorf("remaining subscriber got %v, want RO1", ev.Payload)
	}
}

func TestUnsubscribeIsIdempotent(t *testing.T) {
	eb := NewEventBus()
	ch := eb.Subscribe(RunningOrderUpdated, 1)

	eb.Unsubscribe(RunningOrderUpdated, ch)
	eb.Unsubscribe(RunningOrderUpdated, ch) // must not panic on double close
	eb.Unsubscribe(StoryModified, ch)       // wrong type is ignored

	if n := eb.SubscriberCount(RunningOrderUpdated); n != 0 {
		t.Errorf("SubscriberCount = %d, want 0", n)
	}
}

// Concurrent publish and subscribe/unsubscribe cycles must not leak
// subscribers or send on a closed channel. Run with -race.
func TestSubscribeUnsubscribeCyclesWithConcurrentPublish(t *testing.T) {
	eb := NewEventBus()
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				eb.Publish(Event{Type: RunningOrderUpdated})
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				ch := eb.Subscribe(RunningOrderUpdated, 1)
				eb.Unsubscribe(RunningOrderUpdated, ch)
			}
		}()
	}
	wg.Wait()
	close(done)

	if n := eb.SubscriberCount(RunningOrderUpdated); n != 0 {
		t.Errorf("SubscriberCount after cycles = %d, want 0", n)
	}
}
