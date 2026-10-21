# hookrelay

A small Go service that receives signed payment webhooks and stores each event once.
It's the demo project for a GoLand blog post about Air and the 2026.3 IDE skills.

Standard library only, plus testify in the tests. Requires Go 1.27.3 or later.

## Run the service

```sh
go run .
```

The server listens on `:8080` and accepts `POST /webhooks/{provider}`.
Requests must carry a `Webhook-Signature` header signed with `WEBHOOK_SECRET`. If the variable isn't set, the server uses `dev-secret`.

```sh
WEBHOOK_SECRET=my-secret go run .
```

### Send requests from GoLand

Open `requests/webhooks.http`, select the `dev` environment, and run the requests.
While the [demo issues](#demo-issues) are in place, the valid and duplicate requests return `400 invalid JSON body`. Once they're fixed you should see:

| Request | Expected response |
|---|---|
| Valid event | `202 Accepted` the first time, `200 OK` after that |
| Duplicate event | `200 OK` |
| Bad signature | `401 Unauthorized` |

A pre-request script signs each request with the `secret` from `requests/http-client.env.json`, so there's nothing to compute by hand.

### Signature format

```
Webhook-Signature: t=<unix seconds>,v1=<lowercase hex HMAC-SHA256 of "<t>.<body>">
```

Requests with a timestamp older than five minutes are rejected.

### Errors

Error responses carry a short, fixed message such as `invalid JSON body` or `invalid signature`. The underlying error is logged with `log/slog` and never sent to the client.

## Run the tests

```sh
go test ./...
go test -race ./...
go vet ./...
```

### Benchmark and profile the handler

```sh
go test -run '^$' -bench BenchmarkHandler -benchmem -cpuprofile cpu.out ./internal/webhook
go tool pprof -top -cum cpu.out
```

In GoLand, click the gutter icon next to `BenchmarkHandler` and choose **Profile with CPU Profiler**.

## Demo issues

The code contains deliberate issues for the demo. Each one is marked with an `// ISSUE:` comment that describes it and a `// FIX:` comment that says what should fix it. List them with:

```sh
grep -rn "ISSUE:" --include='*.go' .
```

| Issue | File | Fixed by |
|---|---|---|
| The event ID regex is compiled on every call. It's the top cost in the `BenchmarkHandler` CPU profile. | `internal/webhook/validate.go` | Profiling |
| A duplicate delivery that races past the `Get` check gets `500`, so the provider retries forever. | `internal/webhook/handler.go` | Diff feedback on Agent 1 |
| Signatures are compared with `==`, which isn't constant time. | `internal/middleware/signature.go` | Diff feedback on Agent 2, and Code Review |
| Request bodies are read without `http.MaxBytesReader`. | `internal/webhook/handler.go`, `internal/middleware/signature.go` | Code Review |
| The middleware discards the `io.ReadAll` error. | `internal/middleware/signature.go` | Code Review |
| The middleware reads the body and doesn't put it back, so real requests fail with `400`. `go test ./...` still passes because the middleware test's `next` never reads the body. | `internal/middleware/signature.go`, `internal/middleware/signature_test.go` | Debugger |
| Verification is hard-wired to one HMAC scheme and one secret, so adding a second provider means rewriting the middleware. | `internal/middleware/signature.go` | Refactoring |
