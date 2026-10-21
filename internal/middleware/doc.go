// Package middleware contains HTTP middleware for webhook routes.
//
// VerifySignature reads the request body, checks the Webhook-Signature
// header with a Verifier, and restores the body for the next handler.
// HMACVerifier implements Verifier with HMAC-SHA256 and rejects
// timestamps older than five minutes.
package middleware
