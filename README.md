# Feature-Flag-Service

A feature-flag management service exposed as a REST API in Go, built only on
`net/http` from the standard library. Flags live in a thread-safe in-memory
store guarded by a `sync.RWMutex`; a deterministic rollout decision evaluates
flags for individual users. The service provides input validation, uniform JSON
error objects, access logging as middleware, and full handler tests with
`httptest`.

## Tech Stack

- **Language:** Go (1.22+)
- **Framework:** `net/http` (standard library only — no external web framework)
- **Storage:** in-memory store with `sync.RWMutex`
- **Testing:** `go test` with `net/http/httptest`

## Install / Build

No external dependencies. To build:

```sh
go build ./...
```

To run the tests:

```sh
go test ./...
```

## Run (dev)

The server reads the `PORT` environment variable (default `8080`):

```sh
go run .
```

or with an explicit port:

```sh
PORT=9000 go run .
```

The server then listens on `http://localhost:8080`.

## Endpoints

| Method | Path                    | Description                                      |
| ------ | ----------------------- | ------------------------------------------------ |
| POST   | `/flags`                | Create a flag (201 / 400 / 409 / 413)            |
| GET    | `/flags`                | List all flags, sorted by key (200)              |
| GET    | `/flags/{key}`          | Get a single flag (200 / 404)                    |
| PUT    | `/flags/{key}`          | Update only the supplied fields (200 / 400 / 404 / 413) |
| DELETE | `/flags/{key}`          | Delete a flag (204 / 404)                        |
| GET    | `/flags/{key}/evaluate` | Deterministically evaluate a flag for a user (200 / 400 / 404) |
| GET    | `/healthz`              | Health check (200 `{"status":"ok"}`)             |

### Flag JSON shape

```json
{
  "key": "my-flag",
  "enabled": true,
  "description": "optional description",
  "rollout_percent": 50
}
```

Every JSON response (success and error) sets `Content-Type: application/json`.
Errors use a uniform shape: `{"error": "<generic message>"}`.

## Features

- CRUD for feature flags over a REST API.
- Deterministic rollout evaluation per user.
- Thread-safe in-memory storage.
- Uniform JSON error objects with generic, safe messages.
- Access logging middleware.
- Health endpoint.
