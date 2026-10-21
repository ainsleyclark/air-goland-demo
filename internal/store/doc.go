// Package store persists webhook events, keyed by event ID.
//
// Save returns ErrDuplicate when an event with the same ID already exists,
// so callers can treat redeliveries as processed. Memory is the in-memory
// implementation; other backends only need to implement Store.
package store
