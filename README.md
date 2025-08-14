
# PocketBase Experiments

> Advanced authentication playground with PocketBase, featuring WebAuthn/Passkeys, TOTP, and modern web technologies.

## 🚀 Features

- **🔐 Advanced Authentication**
  - WebAuthn/Passkeys implementation
  - TOTP (Time-based One-Time Passwords)
  - Traditional email/password authentication
  - Multi-factor authentication support

- **🛠️ Modern Tech Stack**
  - **Backend**: Go with PocketBase framework (refactored & modular)
  - **Frontend**: SvelteKit 2.x with TypeScript
  - **Database**: SQLite with PocketBase ORM
  - **Styling**: TailwindCSS + DaisyUI
  - **Build**: Embedded frontend with Go embed
  - **Testing**: Comprehensive Go test suite

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

# Run tests
go test -v

# Run tests with coverage
go test -v -cover

# Update all Go modules
go get -u -t ./...
go mod tidy

# Build and test
go build -o /dev/null . && go test -v
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
├── main.go              # Main application entry point & routing
├── config.go            # Environment configuration management  
├── auth.go              # Authentication service & WebAuthn setup
├── types.go             # Type definitions & interfaces
├── handlers_totp.go     # TOTP-related HTTP handlers
├── handlers_webauthn.go # WebAuthn-related HTTP handlers
├── utils.go             # Utility functions
├── models.go            # User models & WebAuthn interface
├── store.go             # In-memory session store
├── core_test.go         # Comprehensive test suite
├── ui/                  # SvelteKit frontend
│   ├── src/routes/      # Application routes
│   ├── src/lib/         # Components and utilities
│   └── embed.go         # Frontend embedding
└── pb_data/             # PocketBase database and storage
```

### Database Collections
- **users**: User accounts with TOTP secrets
- **credentials**: WebAuthn credentials
- **_mfas**: Multi-factor authentication records

### Code Architecture & Quality

**🔧 Refactored Codebase:**
- **Modular Design**: Separated concerns into dedicated files for better maintainability
- **Clean Architecture**: Service layer pattern with dependency injection
- **Type Safety**: Comprehensive type definitions and interfaces
- **Error Handling**: Standardized error responses and proper error propagation

**🧪 Testing Suite:**
- **Unit Tests**: Comprehensive test coverage for all core modules
- **Edge Cases**: Tests for invalid inputs, missing data, and error conditions  
- **Data Integrity**: Serialization/deserialization round-trip testing
- **Security**: Session ID uniqueness and cryptographic validation
- **Configuration**: Environment setup and validation testing

**Test Coverage:**
```bash
# Run the full test suite
go test -v

# Key test areas:
✅ Configuration management & environment handling
✅ Authentication service & WebAuthn initialization
✅ Session management with secure session IDs
✅ Data structures & JSON serialization
✅ Base64 URL encoding for WebAuthn compliance
✅ Error handling & edge cases
```

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

**Architecture:**
- **Modular Backend**: Refactored Go code with clear separation of concerns
- **Service Layer**: Authentication service with dependency injection pattern
- **Handler Organization**: Separate handlers for TOTP and WebAuthn functionality
- **Configuration Management**: Environment-based config with validation

**Technical Details:**
- The frontend is automatically built and embedded into the Go binary via `go:generate`
- WebAuthn sessions are stored in-memory (consider Redis for production scaling)
- PocketBase handles user management, while custom routes handle advanced auth
- Static files are served with gzip compression
- SPA fallback ensures proper routing for client-side navigation

**Quality Assurance:**
- Comprehensive test suite ensures code reliability
- All core functionality is unit tested
- Error handling is standardized across the application
- Session management includes cryptographic security validation

## 📚 API Documentation

All custom authentication endpoints require proper headers and CORS handling. WebAuthn endpoints use session tokens for state management during the authentication ceremony.

For detailed API usage examples, see the frontend implementation in `ui/src/lib/components/`.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. **Run tests**: `go test -v` (backend) and `npm test` (frontend)
5. **Run linting**: `go vet ./...` (backend) and `npm run lint` (frontend)
6. Ensure all tests pass and code follows project patterns
7. Submit a pull request

### Development Workflow
```bash
# Backend development
go mod tidy                    # Install dependencies
go test -v                     # Run tests
go build -o /dev/null .        # Verify build
go run . serve                 # Start development server

# Frontend development (in ui/ directory)
npm install                    # Install dependencies  
npm run dev                    # Start dev server with hot reload
npm test                       # Run frontend tests
npm run lint                   # Check code style
```

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
