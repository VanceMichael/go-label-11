package event

import (
	"fmt"
	"strings"

	"github.com/VanceMichael/go-base-airbridge/internal/domain"
)

// Delivery is the operator-facing representation sent to live flight feeds.
// Headers and Checkpoints are intentionally exposed because clients attach
// local presentation metadata while rendering an event.
type Delivery struct {
	Envelope    Envelope
	Headers     map[string]string
	Checkpoints []string
}

func NewDelivery(envelope Envelope, headers map[string]string, checkpoints []string) (Delivery, error) {
	if err := envelope.Validate(); err != nil {
		return Delivery{}, err
	}
	if len(checkpoints) == 0 {
		return Delivery{}, fmt.Errorf("%w: delivery checkpoints", domain.ErrInvalid)
	}
	for _, checkpoint := range checkpoints {
		if strings.TrimSpace(checkpoint) == "" {
			return Delivery{}, fmt.Errorf("%w: blank delivery checkpoint", domain.ErrInvalid)
		}
	}
	for key := range headers {
		if strings.TrimSpace(key) == "" {
			return Delivery{}, fmt.Errorf("%w: blank delivery header", domain.ErrInvalid)
		}
	}
	return Delivery{
		Envelope:    envelope,
		Headers:     headers,
		Checkpoints: checkpoints,
	}, nil
}

// Snapshot returns a value suitable for handing to another feed consumer.
func (d Delivery) Snapshot() Delivery {
	return Delivery{
		Envelope:    d.Envelope,
		Headers:     d.Headers,
		Checkpoints: d.Checkpoints,
	}
}

func (d Delivery) Validate() error {
	if err := d.Envelope.Validate(); err != nil {
		return err
	}
	if len(d.Checkpoints) == 0 {
		return fmt.Errorf("%w: delivery checkpoints", domain.ErrInvalid)
	}
	for _, checkpoint := range d.Checkpoints {
		if strings.TrimSpace(checkpoint) == "" {
			return fmt.Errorf("%w: blank delivery checkpoint", domain.ErrInvalid)
		}
	}
	return nil
}

func (d Delivery) TenantID() string {
	return d.Envelope.TenantID
}

func (d Delivery) EventID() string {
	return d.Envelope.ID
}
