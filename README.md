# go-starter-auth-app

Reusable starter application for authenticated server-rendered Go projects.

`go-starter-auth-app` wires together my reusable Go web modules into a working application foundation with login, sessions, protected routes, logout, SQLite persistence, security middleware, and server-rendered templates.

It is designed to be used as a GitHub template repository for new authenticated Go web projects.

## Built with

* [`go-web-core`](https://github.com/danieljmanningdev/go-web-core)
* [`go-web-security`](https://github.com/danieljmanningdev/go-web-security)
* [`go-web-auth`](https://github.com/danieljmanningdev/go-web-auth)

## Features

* Go `net/http`
* SQLite database
* SQL migrations
* Structured logging
* Request middleware
* Server-side HTML rendering
* Login with bcrypt password verification
* Secure session-token generation
* Hashed session tokens in the database
* HTTP-only session cookies
* Protected routes
* Logout with session invalidation
* Cross-origin request protection
* Secure response headers
* Panic recovery
* Development and production configuration
* Automated vulnerability scanning support

## Current routes

```text
/login       GET, POST
/dashboard   GET - authenticated only
/logout      POST
```

### `/login`

`GET /login` renders the login form.

`POST /login`:

1. Looks up the user by email.
2. Verifies the supplied password.
3. Generates a secure session token.
4. Stores the hashed token in SQLite.
5. Sets the raw token in an HTTP-only cookie.
6. Redirects to `/dashboard`.

### `/dashboard`

Protected by authentication middleware.

Anonymous requests are redirected to:

```text
/login
```

### `/logout`

`POST /logout`:

1. Reads the current session token.
2. Hashes the token.
3. Deletes the matching database session.
4. Clears the browser session cookie.
5. Redirects to `/login`.

## Project structure

```text
.
├── cmd
│   └── server
│       └── main.go
├── internal
│   └── auth
│       └── store.go
├── migrations
│   └── 001_auth.sql
├── web
│   └── templates
│       ├── dashboard.html
│       └── login.html
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

Runtime SQLite data should not be committed to the repository.

For example:

```gitignore
data/
```

## Database

The starter uses SQLite.

The initial migration creates:

```text
users
sessions
schema_migrations
```

The `users` table stores:

```text
id
email
password_hash
created_at
```

The `sessions` table stores:

```text
id
user_id
token_hash
expires_at
created_at
```

Raw session tokens are not stored in the database.

The browser receives the raw session token, while the database stores its SHA-256 hash.

## Authentication flow

```text
email + password
       ↓
find user
       ↓
verify bcrypt hash
       ↓
generate random session token
       ↓
hash session token
       ↓
store hash in SQLite
       ↓
set raw token as HTTP-only cookie
       ↓
redirect to protected application
```

## Running locally

Clone the repository or create a new repository using this GitHub template.

Install/update dependencies:

```bash
go mod tidy
```

Run the checks:

```bash
gofmt -w .
go test ./...
go vet ./...
govulncheck ./...
git diff --check
```

Start the application:

```bash
go run ./cmd/server
```

The default server address is:

```text
http://localhost:8080
```

## Creating the first user

The starter deliberately does not include public registration.

Applications can add their own:

* registration flow
* administrator-created users
* invitations
* account provisioning
* external identity provider

For development, generate a bcrypt password hash using `go-web-auth/password`.

Example:

```go
package main

import (
	"fmt"
	"log"

	"github.com/danieljmanningdev/go-web-auth/password"
)

func main() {
	hash, err := password.Hash("change-me")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(hash)
}
```

Insert the generated hash into the development database:

```sql
INSERT INTO users (
    email,
    password_hash
)
VALUES (
    'demo@example.com',
    'YOUR_BCRYPT_HASH'
);
```

Do not use example credentials or passwords in production.

## Configuration

Configuration is provided through `go-web-core/config`.

Supported environment variables include:

```text
APP_ENV
APP_PORT
LOG_LEVEL
DATABASE_PATH
TEMPLATE_DIR
```

Defaults:

```text
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=info
DATABASE_PATH=./data/app.db
TEMPLATE_DIR=web/templates
```

Session cookies are configured as secure by default.

For non-production environments, this starter disables the `Secure` cookie flag so authentication can work over local HTTP.

Production applications should use HTTPS.

## Security

The starter uses `go-web-security` for:

* cross-origin request protection
* defensive HTTP response headers
* panic recovery

It uses `go-web-auth` for:

* bcrypt password hashing
* secure random session tokens
* SHA-256 session-token hashing
* HTTP-only session cookies
* authentication middleware
* protected routes

Security checks can be run with:

```bash
govulncheck ./...
```

and optionally:

```bash
gosec ./...
```

## Using this as a template

This repository is intended to be marked as a GitHub **Template repository**.

For a new project:

1. Click **Use this template** on GitHub.
2. Create the new project repository.
3. Clone the new repository.
4. Update the module path.
5. Run `go mod tidy`.
6. Add project-specific migrations, templates, handlers, and business logic.

Update the Go module path with:

```bash
go mod edit -module=github.com/YOUR-USERNAME/YOUR-PROJECT
```

Then:

```bash
go mod tidy
```

The reusable modules remain versioned dependencies, so fixes and improvements can be pulled into future projects without copying their source code.

## What belongs in the starter

This repository contains reusable application wiring.

It should remain focused on:

```text
application startup
database bootstrap
authentication wiring
session handling
security middleware
protected routing
basic templates
```

Client-specific functionality should be added in the project created from this template.

Examples include:

```text
clients
projects
contracts
bookings
billing
documents
uploads
notifications
business-specific workflows
```

## Design philosophy

The reusable stack is split into two layers.

### Reusable modules

```text
go-web-core
go-web-security
go-web-auth
```

These contain shared implementation.

### Starter application

```text
go-starter-auth-app
```

This wires those modules together into a ready-to-extend project structure.

The goal is:

> Start new authenticated Go projects with the infrastructure already working and spend development time on the client-specific product instead.

## Development

Format:

```bash
gofmt -w .
```

Test:

```bash
go test ./...
```

Static analysis:

```bash
go vet ./...
```

Vulnerability scan:

```bash
govulncheck ./...
```

Whitespace check:

```bash
git diff --check
```

## Status

This project is an early reusable starter and may evolve as it is used across real projects.

## License

See [LICENSE](LICENSE).
