package store

import (
	"context"
	"errors"
)

var (
	// ErrDuplicate is returned by Save when an event with the same ID exists.
	ErrDuplicate = errors.New("store: duplicate event")

	// ErrNotFound is returned by Get when no event has the given ID.
	ErrNotFound = errors.New("store: event not found")
)

// Store saves webhook events, keyed by event ID.
// Implementations must be safe for concurrent use.
type Store interface {
	// Get returns the event with the given ID.
	//
	// Returns ErrNotFound when no event has that ID.
	Get(ctx context.Context, id string) (Event, error)

	// Save stores ev under its ID.
	//
	// Returns ErrDuplicate when an event with the same ID exists.
	Save(ctx context.Context, ev Event) error
}
