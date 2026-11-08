package server

import (
	"context"
	"net"
	"testing"
	"time"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/events"
)

func newTestServer(bus *events.EventBus) (*TCPServer, *config.Config) {
	cfg := &config.Config{}
	cfg.MOS.ID = "test.mos"
	cfg.MOS.HeartbeatInterval = time.Minute
	cfg.MOS.ClientTimeout = time.Minute

	return &TCPServer{
		clients:    make(map[string]*ClientConnection),
		config:     cfg,
		eventBus:   bus,
		shutdownCh: make(chan struct{}),
	}, cfg
}

func waitForSubscribers(t *testing.T, bus *events.EventBus, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for bus.SubscriberCount(events.RunningOrderUpdated) != want {
		if time.Now().After(deadline) {
			t.Fatalf("SubscriberCount = %d, want %d",
				bus.SubscriberCount(events.RunningOrderUpdated), want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// Regression test for MOS-10: each disconnected client must release its
// event bus subscription.
func TestClientConnectDisconnectCyclesReleaseSubscriptions(t *testing.T) {
	bus := events.NewEventBus()
	srv, cfg := newTestServer(bus)

	for i := 0; i < 5; i++ {
		serverSide, clientSide := net.Pipe()
		client := NewClientConnection(serverSide, srv, cfg)

		done := make(chan struct{})
		go func() {
			client.Start(context.Background())
			close(done)
		}()
		waitForSubscribers(t, bus, 1)

		// Remote end hangs up: the read loop sees EOF and Start returns
		clientSide.Close()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("cycle %d: client.Start did not return after disconnect", i)
		}
		waitForSubscribers(t, bus, 0)
	}
}

func TestClientCloseReleasesSubscription(t *testing.T) {
	bus := events.NewEventBus()
	srv, cfg := newTestServer(bus)

	serverSide, clientSide := net.Pipe()
	defer clientSide.Close()
	client := NewClientConnection(serverSide, srv, cfg)

	done := make(chan struct{})
	go func() {
		client.Start(context.Background())
		close(done)
	}()
	waitForSubscribers(t, bus, 1)

	// Server-initiated close, as the heartbeat timeout does
	client.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("client.Start did not return after Close")
	}
	waitForSubscribers(t, bus, 0)
}
