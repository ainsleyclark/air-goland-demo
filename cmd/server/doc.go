// Server runs the hookrelay webhook receiver on :8080.
//
// It serves POST /webhooks/{provider}, verifies each request's signature
// with WEBHOOK_SECRET (or a dev secret when unset), and keeps events in
// memory, so they are lost on restart.
package main
