
# PocketBase Experiments

> Advanced authentication playground with PocketBase, featuring WebAuthn/Passkeys, TOTP, and modern web technologies.

## 🚀 Features

- **🔐 Advanced Authentication**
  - WebAuthn/Passkeys implementation
  - TOTP (Time-based One-Time Passwords)
  - Traditional email/password authentication
  - Multi-factor authentication support

- **🛠️ Modern Tech Stack**
  - **Backend**: Go with PocketBase framework
  - **Frontend**: SvelteKit 2.x with TypeScript
  - **Database**: SQLite with PocketBase ORM
  - **Styling**: TailwindCSS + DaisyUI
  - **Build**: Embedded frontend with Go embed

- **📱 User Experience**
  - Responsive design with mobile-first approach
  - SPA routing with fallback support
  - Marketing pages and admin dashboard
  - Account management and security settings

## 📋 Prerequisites

- Go 1.24.0 or higher
- Node.js 18+ and npm
- Git

## ⚡ Quick Start

### 1. Clone and Install

```bash
git clone <repository-url>
cd pocketbase-experiments
go mod tidy
```

### 2. Environment Setup

Create `.env` file for development:
```bash
TOTP_ISSUER="PocketBase Experiments"
PROTO="http"
HOST="localhost"
PORT=":8090"
```

For production, create `.env.production` with appropriate values.

### 3. Frontend Setup

```bash
cd ui
npm install
cd ..
```

### 4. Run Development Server

```bash
go run . serve
```

The application will be available at `http://localhost:8090`

## 🛠️ Development

### Backend Commands

```bash
# Install/update dependencies
go mod tidy

# Run development server
go run . serve

# Update all Go modules
go get -u -t ./...
go mod tidy
```

### Frontend Commands (ui/ directory)

```bash
# Install dependencies
npm install

# Development server (with hot reload)
npm run dev

# Build for production
npm run build

# Type checking
npm run check
npm run check:watch

# Testing
npm run test
npm run test_run

# Linting and formatting
npm run lint
npm run format
npm run format_check
```

## 🔐 Authentication Features

### WebAuthn/Passkeys
- Passwordless authentication using FIDO2/WebAuthn standard
- Support for Touch ID, Face ID, security keys, and platform authenticators
- Secure credential storage in SQLite database

**API Endpoints:**
- `POST /api/pb-experiments/passkey/registerStart` - Begin passkey registration
- `POST /api/pb-experiments/passkey/registerFinish` - Complete passkey registration
- `POST /api/pb-experiments/passkey/loginStart` - Begin passkey authentication
- `POST /api/pb-experiments/passkey/loginFinish` - Complete passkey authentication

### TOTP (Time-based OTP)
- QR code generation for authenticator apps
- Support for Google Authenticator, Authy, etc.
- Backup codes and account recovery

**API Endpoints:**
- `GET /api/pb-experiments/get-qr` - Generate TOTP QR code
- `POST /api/pb-experiments/totp-login` - Verify TOTP passcode

## 🏗️ Architecture

### Project Structure
```
├── main.go              # Main application and API routes
├── models.go            # User models and WebAuthn interface
├── store.go            # In-memory session store
├── ui/                 # SvelteKit frontend
│   ├── src/routes/     # Application routes
│   ├── src/lib/        # Components and utilities
│   └── embed.go        # Frontend embedding
└── pb_data/            # PocketBase database and storage
```

### Database Collections
- **users**: User accounts with TOTP secrets
- **credentials**: WebAuthn credentials
- **_mfas**: Multi-factor authentication records

## 🚀 Production Deployment

### Build for Production

```bash
# Generate and build frontend, then compile Go binary
go generate ./...
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o pocketbase-experiments
```

### Run Production Server

```bash
# Ensure .env.production exists with proper configuration
./pocketbase-experiments serve
```

### Environment Variables (Production)

```bash
TOTP_ISSUER="Your App Name"
PROTO="https"
HOST="yourdomain.com"
# PORT not needed for production (no :port suffix)
```

## 🔧 Development Notes

- The frontend is automatically built and embedded into the Go binary via `go:generate`
- WebAuthn sessions are stored in-memory (consider Redis for production scaling)
- PocketBase handles user management, while custom routes handle advanced auth
- Static files are served with gzip compression
- SPA fallback ensures proper routing for client-side navigation

## 📚 API Documentation

All custom authentication endpoints require proper headers and CORS handling. WebAuthn endpoints use session tokens for state management during the authentication ceremony.

For detailed API usage examples, see the frontend implementation in `ui/src/lib/components/`.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## 📄 License

This project is open source. Please check the license file for details.
