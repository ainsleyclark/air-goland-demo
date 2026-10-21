# hookrelay

A small Go service that receives signed payment webhooks and stores each event once.
It's the demo project for a GoLand blog post about Air and the 2026.3 IDE skills.

Standard library only. Requires Go 1.27.3 or later.

## Run the service

```sh
go run ./cmd/server
```

The server listens on `:8080` and accepts `POST /webhooks/{provider}`.
Requests must carry a `Webhook-Signature` header signed with `WEBHOOK_SECRET`. If the variable isn't set, the server uses `dev-secret`.

```sh
WEBHOOK_SECRET=my-secret go run ./cmd/server
```

### Send requests from GoLand

Open `requests/webhooks.http`, select the `dev` environment, and run the requests:

| Request | Expected response |
|---|---|
| Valid event | `202 Accepted` the first time, `200 OK` after that |
| Duplicate event | `200 OK` |
| Bad signature | `401 Unauthorized` |

A pre-request script signs each request with the `secret` from `requests/http-client.env.json`, so there's nothing to compute by hand.

### Signature format

```
Webhook-Signature: t=<unix seconds>,v1=<hex HMAC-SHA256 of "<t>.<body>">
```

Requests with a timestamp older than five minutes are rejected.

### Errors

Error responses carry a short, fixed message such as `invalid JSON body` or `invalid signature`. The underlying error is logged with `log/slog` and never sent to the client. Bodies larger than 1 MiB get `413 Request Entity Too Large`.

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

## Demo history

The blog post walks through the history of `main`. Two agent commits add the handler and the middleware with deliberate issues, a merge brings them together, and later commits fix each issue. The code has no comments pointing at the issues.

Run `git checkout <commit>` to go back to any of these states:

| Commit | State |
|---|---|
| `b98784a` | Baseline: store, `validateEventID`, and a route stub that returns `501` |
| `c2aed50` | Agent 1 adds the webhook handler |
| `089519c` | Agent 2 adds the signature verification middleware |
| `71598b6` | Both merged, before any fixes. The valid request returns `400 EOF` |

| Issue | Introduced in | Found with | Fixed in |
|---|---|---|---|
| `validateEventID` compiles its regular expression on every call. It's the top cost in the `BenchmarkHandler` CPU profile. | `b98784a`, measured at `c2aed50` | Profiling | `6d61d6c`: compile it once in a package-level `var` |
| The handler returns `500` for any `Save` error, including `store.ErrDuplicate`. Two deliveries of the same event can both pass the `Get` check; the second `Save` fails and the provider keeps retrying. | `c2aed50` | Diff feedback: "return 200 for duplicates, or the provider retries forever" | `80be426`: return `200` for `ErrDuplicate`, and drop the `Get` check |
| The handler decodes JSON straight from `r.Body` with no size limit. | `c2aed50` | Code Review | `6d1a177`: wrap the body in `http.MaxBytesReader`, in the handler and the middleware |
| The middleware reads the body with `io.ReadAll` and doesn't put it back. After the merge, the valid request returns `400 EOF`, but `go test ./...` still passes because the middleware tests use a `next` handler that never reads the body. | `089519c`, visible at `71598b6` | Debugger | `f1f9324`: restore `r.Body` before calling `next` |
| The middleware compares signatures with `==`, which isn't constant time. | `089519c` | Code Review | `41118f1`: decode the hex signature and compare with `hmac.Equal` |
| The handler and middleware send raw Go errors to the client, such as JSON decode errors and `EOF`, and log nothing. | `c2aed50`, `089519c` | Code Review | `942e021`: log the error with `slog` and respond with a fixed message |

After the fixes, `fb6935c` moves signature checking behind a `Verifier` interface with one HMAC implementation, and `b56c2dc` adds an end-to-end test that posts a signed event twice through the middleware.
