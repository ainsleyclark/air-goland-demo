// Package webhook receives payment provider webhooks over HTTP.
//
// Handler decodes each event, validates its ID, and saves it to a
// store.Store. Saving is idempotent: a new event gets 202 Accepted and a
// redelivered one gets 200 OK, so providers stop retrying. Signatures are
// checked before the handler runs, by the middleware package.
package webhook
