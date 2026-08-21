package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/VanceMichael/go-base-airbridge/internal/domain"
	"github.com/VanceMichael/go-base-airbridge/internal/event"
)

// ReplayBus fans flight updates out to operator consoles and keeps the most
// recent update per tenant for consoles that reconnect after an interruption.
type ReplayBus struct {
	mu          sync.RWMutex
	next        int
	subscribers map[int]replaySubscription
	latest      map[string]event.Delivery
}

type replaySubscription struct {
	tenantID string
	updates  chan event.Delivery
}

func NewReplayBus() *ReplayBus {
	return &ReplayBus{
		subscribers: make(map[int]replaySubscription),
		latest:      make(map[string]event.Delivery),
	}
}

func (b *ReplayBus) Subscribe(tenantID string, buffer int) (int, <-chan event.Delivery, error) {
	if tenantID == "" || buffer < 1 {
		return 0, nil, domain.ErrInvalid
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.next
	b.next++
	updates := make(chan event.Delivery, buffer)
	b.subscribers[id] = replaySubscription{tenantID: tenantID, updates: updates}
	return id, updates, nil
}

func (b *ReplayBus) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subscription, ok := b.subscribers[id]
	if !ok {
		return
	}
	delete(b.subscribers, id)
	close(subscription.updates)
}

func (b *ReplayBus) Publish(ctx context.Context, delivery event.Delivery) error {
	if err := delivery.Validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.latest[delivery.TenantID()] = delivery.Snapshot()
	for _, subscription := range b.subscribers {
		if subscription.tenantID != delivery.TenantID() {
			continue
		}
		select {
		case subscription.updates <- delivery.Snapshot():
		case <-ctx.Done():
			return fmt.Errorf("publish flight update: %w", ctx.Err())
		}
	}
	return nil
}

func (b *ReplayBus) Latest(tenantID string) (event.Delivery, error) {
	if tenantID == "" {
		return event.Delivery{}, domain.ErrInvalid
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	delivery, ok := b.latest[tenantID]
	if !ok {
		return event.Delivery{}, domain.ErrNotFound
	}
	return delivery.Snapshot(), nil
}

func (b *ReplayBus) SubscriberCount(tenantID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	count := 0
	for _, subscription := range b.subscribers {
		if subscription.tenantID == tenantID {
			count++
		}
	}
	return count
}
