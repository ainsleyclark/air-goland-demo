# hookrelay

A small Go service that receives signed payment webhooks and stores each event once.
It's the demo project for a GoLand blog post about Air and the 2026.3 IDE skills, and it contains a few deliberate issues (see [Demo notes](#demo-notes)).

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

Open `requests/webhooks.http`, select the `dev` environment, and run the requests. On `final` you should see:

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

## Branches

| Branch | Contents |
|---|---|
| `main` | Baseline: store, `validateEventID`, and a route stub that returns `501` |
| `agent-1` | The webhook handler |
| `agent-2` | The signature verification middleware |
| `merged` (tag) | `agent-1` and `agent-2` merged, before any fixes |
| `final` | The merge plus every fix |

## Demo notes

These issues are deliberate. The code has no comments pointing at them.

| Issue | Branch | Found with | Fix on `final` |
|---|---|---|---|
| `validateEventID` compiles its regular expression on every call. It's the top cost in the `BenchmarkHandler` CPU profile. | `main`, measured on `agent-1` | Profiling | Compile it once in a package-level `var` |
| The handler returns `500` for any `Save` error, including `store.ErrDuplicate`. Two deliveries of the same event can both pass the `Get` check; the second `Save` fails and the provider keeps retrying. | `agent-1` | Diff feedback: "return 200 for duplicates, or the provider retries forever" | Return `200` for `ErrDuplicate`, and drop the `Get` check |
| The handler decodes JSON straight from `r.Body` with no size limit. | `agent-1` | Code Review | Wrap the body in `http.MaxBytesReader`, in the handler and the middleware |
| The middleware reads the body with `io.ReadAll` and doesn't put it back. After the merge, the valid request returns `400 EOF`, but `go test ./...` still passes because the middleware tests use a `next` handler that never reads the body. | `agent-2`, visible at `merged` | Debugger | Restore `r.Body` before calling `next` |
| The middleware compares signatures with `==`, which isn't constant time. | `agent-2` | Code Review | Decode the hex signature and compare with `hmac.Equal` |
| The handler and middleware send raw Go errors to the client, such as JSON decode errors and `EOF`, and log nothing. | `agent-1`, `agent-2` | Code Review | Log the error with `slog` and respond with a fixed message, such as `invalid JSON body` or `invalid signature`. Oversized bodies get `413` |

`final` also moves signature checking behind a `Verifier` interface with one HMAC implementation, adds an end-to-end test that posts a signed event twice through the middleware, and is clean under `go test -race ./...` and `go vet ./...`.
