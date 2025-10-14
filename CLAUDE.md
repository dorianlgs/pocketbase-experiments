# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a PocketBase-based application that combines a Go backend with a SvelteKit frontend. The project implements advanced authentication features including passkeys/WebAuthn, TOTP (Time-based One-Time Passwords), and traditional email/password authentication.

**Tech Stack:**
- **Backend**: Go with PocketBase framework, SQLite database
- **Frontend**: SvelteKit 2.x with TypeScript, TailwindCSS, DaisyUI
- **Authentication**: WebAuthn/Passkeys, TOTP, traditional auth
- **Build**: Vite, embedded frontend via Go embed

## Development Commands

### Go Backend
```bash
# Install dependencies
go mod tidy

# Development server
go run . serve

# Build for production
go generate ./...
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"

# Update all modules
go get -u -t ./...
go mod tidy
```

### Frontend (ui/ directory)
```bash
# Install dependencies
npm install

# Development server
npm run dev

# Build for production
npm run build

# Type checking
npm run check
npm run check:watch

# Testing
npm run test
npm run test_run

# Linting & formatting
npm run lint
npm run format
npm run format_check
```

## Architecture

### Backend Structure (Go)
- **main.go**: Main application entry point, PocketBase setup, custom API routes for passkeys and TOTP
- **models.go**: User model implementing WebAuthn interface, credential management
- **store.go**: In-memory session store for WebAuthn flows, user creation/retrieval
- **ui/embed.go**: Embeds built frontend into Go binary with `go:generate` directives

### Frontend Structure (SvelteKit)
- **src/routes/(marketing)/**: Public pages (landing, blog, pricing, login)
- **src/routes/(admin)/**: Protected admin interface (account, settings, billing)
- **src/lib/components/**: Reusable Svelte components (GitHubButton, GoogleButton, PasskeyButton, etc.)
- **src/lib/pocketbase.ts**: PocketBase client initialization
- **src/lib/stores/**: Svelte stores for state management
- **src/config.ts**: Application configuration constants

### Authentication Flow
1. **Traditional Auth**: Email/password via PocketBase built-in auth
2. **Passkeys**: Custom WebAuthn implementation with registration/login endpoints:
   - `/api/pb-experiments/passkey/registerStart`
   - `/api/pb-experiments/passkey/registerFinish`
   - `/api/pb-experiments/passkey/loginStart`
   - `/api/pb-experiments/passkey/loginFinish`
3. **TOTP**: Custom endpoints for QR code generation and validation:
   - `/api/pb-experiments/get-qr`
   - `/api/pb-experiments/totp-login`

### Database Collections
- **users**: Standard PocketBase users with additional `totpSecret` field
- **credentials**: WebAuthn credentials linked to users
- **_mfas**: Multi-factor authentication records

## Key Implementation Notes

### WebAuthn Integration
The application implements a full WebAuthn flow using the `github.com/go-webauthn/webauthn` library. Sessions are managed in-memory with cryptographically secure tokens.

### Frontend Build Integration
The frontend is automatically built and embedded into the Go binary using `go:generate` directives in `ui/embed.go`. Running `go generate ./...` will execute `npm install` and `npm run build`.

### Environment Configuration
- Development: Loads `.env` file
- Production: Loads `.env.production` file
- Required env vars: `TOTP_ISSUER`, `PROTO`, `HOST`, `PORT` (dev only)

### Route Structure
- SvelteKit uses grouped routes: `(marketing)` for public pages, `(admin)` for protected areas
- Static adapter configured for deployment as static files
- Fallback to `404.html` for SPA routing

## Security

### Security Documentation
The project maintains comprehensive security documentation in **SECURITY.md**:
- Vulnerability reporting process
- Known security issues and mitigations
- Security best practices for authentication, database, network, and file uploads
- Production deployment security checklist
- Dependency management guidelines

### Known Vulnerabilities

**CVE-2023-36308 (github.com/disintegration/imaging)**
- **Status**: Known, mitigated (no patch available)
- **Impact**: Low - TIFF panic vulnerability in indirect dependency via PocketBase
- **Mitigation**: MIME type restrictions in `pb_schema.json` prevent TIFF uploads
- **Allowed formats**: JPEG, PNG, SVG, GIF, WebP only
- **Details**: See SECURITY.md for full analysis

### File Upload Security
All file upload fields have explicit MIME type restrictions defined in `pb_schema.json`:
- Avatar field (users collection): Limited to image/jpeg, image/png, image/svg+xml, image/gif, image/webp
- No TIFF support in any collection (mitigates CVE-2023-36308)

### Session Security
- WebAuthn sessions use cryptographically secure random token generation
- Session IDs are unique and validated in tests (see `core_test.go`)
- Auth tokens configurable per collection (default: 7 days for users)
- Password reset tokens expire after 30 minutes