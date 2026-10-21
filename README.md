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

## What the agents build

The Air part of the demo runs two agents in parallel. Each builds one piece of the service on top of the store and `validateEventID`, and the handler and middleware in this repo stand in for their output.

| Agent | Builds | Files |
|---|---|---|
| Agent 1 | The webhook handler. It decodes the event from the request body, validates its ID, and saves it once. A new event gets `202 Accepted`, and a redelivery gets `200 OK`, including one that loses a race to save. | `internal/webhook/handler.go`, `internal/webhook/handler_test.go` |
| Agent 2 | The signature middleware. It checks the `Webhook-Signature` header against an HMAC-SHA256 of `<t>.<body>`, compares the two in constant time with `hmac.Equal`, and rejects timestamps older than five minutes. Rejected requests get `401 Unauthorized`. It also wraps the webhook route in `main.go`. | `internal/middleware/signature.go`, `internal/middleware/signature_test.go`, `main.go` |

## Demo issues

The code contains deliberate issues for the demo. Each one is marked with three comments: `// ISSUE:` describes it, `// DEMO:` names the technology that finds it, and `// FIX:` says what to change. List them with:

```sh
grep -rn -A2 "ISSUE:" --include='*.go' .
```

| Issue | File | Demo |
|---|---|---|
| The event ID regex is compiled on every call. It's the top cost in the `BenchmarkHandler` CPU profile. | `internal/webhook/validate.go` | Profiling Skill |
| Request bodies are read without `http.MaxBytesReader`. | `internal/webhook/handler.go`, `internal/middleware/signature.go` | Code Review Skill |
| The middleware discards the `io.ReadAll` error. | `internal/middleware/signature.go` | Code Review Skill |
| The middleware reads the body and doesn't put it back, so real requests fail with `400`. `go test ./...` still passes because the middleware test's `next` never reads the body. | `internal/middleware/signature.go`, `internal/middleware/signature_test.go` | Debugger Skill |
| `verify` only checks one HMAC scheme but its name doesn't say so, so a second provider's scheme has no obvious home. | `internal/middleware/signature.go` | Refactoring Skill |

### Fixed versions

Each file with issues has a `_fixed.go` twin: the same file with the markers removed and the fixes applied, and nothing else changed. The Refactoring issue is the exception, because the rename is left to the skill. The two are swapped with the `fixed` build tag: a normal build uses the files with issues, and `-tags fixed` uses the twins and adds the `_fixed_test.go` tests, which check each fix.

```sh
go run -tags fixed .
go test -tags fixed ./...
go test -tags fixed -run '^$' -bench BenchmarkHandler -benchmem ./internal/webhook
```

To see a fix as a diff, compare a file with its twin, for example `internal/webhook/handler.go` with `internal/webhook/handler_fixed.go`. To work on the fixed code in GoLand, add `fixed` to the custom tags under **Settings | Go | Build Tags**.
