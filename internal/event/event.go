package event

import (
	"encoding/json"
	"github.com/VanceMichael/go-base-airbridge/internal/domain"
	"time"
)

type Envelope struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	TenantID    string          `json:"tenant_id"`
	AggregateID string          `json:"aggregate_id"`
	OccurredAt  time.Time       `json:"occurred_at"`
	Payload     json.RawMessage `json:"payload"`
}

func New(id, typ, tenant, aggregate string, payload any, at time.Time) (Envelope, error) {
	if id == "" || typ == "" || tenant == "" || aggregate == "" || at.IsZero() {
		return Envelope{}, domain.ErrInvalid
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{ID: id, Type: typ, TenantID: tenant, AggregateID: aggregate, OccurredAt: at, Payload: raw}, nil
}
func (e Envelope) Decode(dst any) error {
	if len(e.Payload) == 0 {
		return domain.ErrInvalid
	}
	return json.Unmarshal(e.Payload, dst)
}
func (e Envelope) Validate() error {
	if e.ID == "" || e.Type == "" || e.TenantID == "" || e.AggregateID == "" || e.OccurredAt.IsZero() {
		return domain.ErrInvalid
	}
	if !json.Valid(e.Payload) {
		return domain.ErrInvalid
	}
	return nil
}

// snapshot returns a copy of the envelope whose Payload slice is backed by
// its own byte array, so callers cannot mutate the original bytes through
// either copy.
func (e Envelope) snapshot() Envelope {
	return Envelope{
		ID:          e.ID,
		Type:        e.Type,
		TenantID:    e.TenantID,
		AggregateID: e.AggregateID,
		OccurredAt:  e.OccurredAt,
		Payload:     append(json.RawMessage(nil), e.Payload...),
	}
}
