// Package store persists webhook events.
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
type Store interface {
	Get(ctx context.Context, id string) (Event, error)
	Save(ctx context.Context, ev Event) error
}
