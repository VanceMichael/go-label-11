package stream

import (
	"context"
	"testing"
	"time"

	"github.com/VanceMichael/go-base-airbridge/internal/event"
)

func TestSubscriberMutationDoesNotPolluteFanoutOrReplay(t *testing.T) {
	envelope, err := event.New(
		"evt-11",
		"shipment.checkpointed",
		"tenant-east",
		"shipment-11",
		map[string]string{"status": "screened"},
		time.Date(2026, 8, 21, 2, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("new envelope: %v", err)
	}
	delivery, err := event.NewDelivery(
		envelope,
		map[string]string{"priority": "normal", "station": "PVG"},
		[]string{"accepted", "screened"},
	)
	if err != nil {
		t.Fatalf("new delivery: %v", err)
	}

	bus := NewReplayBus()
	firstID, firstFeed, err := bus.Subscribe("tenant-east", 1)
	if err != nil {
		t.Fatalf("subscribe first console: %v", err)
	}
	defer bus.Unsubscribe(firstID)
	secondID, secondFeed, err := bus.Subscribe("tenant-east", 1)
	if err != nil {
		t.Fatalf("subscribe second console: %v", err)
	}
	defer bus.Unsubscribe(secondID)

	if err := bus.Publish(context.Background(), delivery); err != nil {
		t.Fatalf("publish: %v", err)
	}
	first := <-firstFeed
	second := <-secondFeed

	first.Headers["priority"] = "locally-highlighted"
	first.Checkpoints[0] = "locally-collapsed"
	first.Envelope.Payload[0] = '!'

	assertDeliveryOriginal(t, "second console", second)
	replayed, err := bus.Latest("tenant-east")
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	assertDeliveryOriginal(t, "retained replay", replayed)

	delivery.Headers["station"] = "CALLER-EDIT"
	delivery.Checkpoints[1] = "CALLER-EDIT"
	delivery.Envelope.Payload[1] = '!'
	replayedAgain, err := bus.Latest("tenant-east")
	if err != nil {
		t.Fatalf("latest after caller edit: %v", err)
	}
	assertDeliveryOriginal(t, "replay after caller edit", replayedAgain)
}

func assertDeliveryOriginal(t *testing.T, label string, delivery event.Delivery) {
	t.Helper()
	if got := delivery.Headers["priority"]; got != "normal" {
		t.Fatalf("%s priority = %q, want normal", label, got)
	}
	if got := delivery.Headers["station"]; got != "PVG" {
		t.Fatalf("%s station = %q, want PVG", label, got)
	}
	if got := delivery.Checkpoints[0]; got != "accepted" {
		t.Fatalf("%s first checkpoint = %q, want accepted", label, got)
	}
	if got := delivery.Checkpoints[1]; got != "screened" {
		t.Fatalf("%s second checkpoint = %q, want screened", label, got)
	}
	var payload map[string]string
	if err := delivery.Envelope.Decode(&payload); err != nil {
		t.Fatalf("%s payload decode: %v", label, err)
	}
	if got := payload["status"]; got != "screened" {
		t.Fatalf("%s payload status = %q, want screened", label, got)
	}
}
