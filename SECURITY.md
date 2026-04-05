# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it by creating a private security advisory on GitHub or by contacting the maintainers directly. Please do not open public issues for security vulnerabilities.

**When reporting, please include:**
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if available)

We will acknowledge receipt of your report within 48 hours and provide a detailed response within 7 days.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < Latest| :x:                |

This is an experimental project. Only the latest version receives security updates.

## Known Security Issues & Mitigations

### CVE-2023-36308: github.com/disintegration/imaging TIFF Vulnerability

**Status:** Known, Mitigated

**Details:**
- **Package:** `github.com/disintegration/imaging` v1.6.2
- **Type:** Indirect dependency (via PocketBase v0.30.2)
- **Vulnerability:** Crafted TIFF files can cause panic via index out of bounds in scanner.go
- **Impact:** Potential denial of service (DoS) through panic
- **CVE:** CVE-2023-36308
- **Upstream Status:** No patch available (last release: November 2019)

**Mitigation:**
This application is **effectively protected** against this vulnerability through input validation:

1. **MIME Type Restrictions:** The avatar upload field explicitly blocks TIFF files
   ```json
   "mimeTypes": [
     "image/jpeg",
     "image/png",
     "image/svg+xml",
     "image/gif",
     "image/webp"
   ]
   ```
   (See: `pb_schema.json` - users collection, avatar field)

2. **No TIFF Support:** No application collections accept TIFF format
3. **PocketBase Validation:** MIME type validation occurs before image processing

**Risk Assessment:** **VERY LOW**
- The vulnerable code path (TIFF processing) cannot be reached
- Input validation prevents malicious TIFF uploads
- Even if exploited, impact is limited to service disruption (not data breach)

**Monitoring:**
- Track PocketBase releases for dependency updates
- Monitor github.com/disintegration/imaging for patches
- Re-evaluate if TIFF support is added to any collection

**Last Reviewed:** 2025-10-14

---

## Security Best Practices

### Authentication & Authorization

1. **WebAuthn/Passkeys**
   - Uses FIDO2/WebAuthn standard for passwordless authentication
   - Credentials stored securely in SQLite database
   - Session tokens use cryptographically secure random generation

2. **TOTP (Time-based OTP)**
   - TOTP secrets stored encrypted in database
   - QR codes generated server-side to prevent client manipulation
   - Supports standard authenticator apps (Google Authenticator, Authy, etc.)

3. **Multi-Factor Authentication (MFA)**
   - MFA can be enabled per user
   - Supports both TOTP and WebAuthn as second factors
   - Session duration: 1800 seconds (30 minutes)

4. **Session Management**
   - WebAuthn sessions stored in-memory with secure session IDs
   - Auth tokens have configurable duration (default: 7 days)
   - Password reset tokens expire after 30 minutes
   - Email verification tokens expire after 3 days

### Database Security

1. **SQLite Configuration**
   - Database stored in `pb_data/data.db`
   - Access controlled by file system permissions
   - Regular backups recommended for production

2. **Sensitive Data**
   - Passwords hashed using bcrypt (PocketBase default)
   - TOTP secrets stored encrypted
   - WebAuthn credentials include only public keys

### Network Security

1. **HTTPS Required**
   - WebAuthn requires HTTPS in production (except localhost)
   - Configure reverse proxy (nginx, Caddy) for TLS termination
   - Use valid SSL/TLS certificates

2. **CORS Configuration**
   - Configure allowed origins appropriately
   - Avoid using wildcard (`*`) in production

3. **API Endpoints**
   - Custom authentication endpoints under `/api/pb-experiments/`
   - Rate limiting recommended for production deployment

### File Upload Security

1. **MIME Type Validation**
   - All file upload fields have explicit MIME type restrictions
   - Avatar uploads limited to: JPEG, PNG, SVG, GIF, WebP
   - File size limits enforced by PocketBase

2. **File Storage**
   - Uploaded files stored in `pb_data/storage/`
   - Files served with proper Content-Type headers
   - Thumbnail generation handled by PocketBase

### Environment Variables

**Development (.env):**
```bash
TOTP_ISSUER="PocketBase Experiments"
PROTO="http"
HOST="localhost"
PORT=":8090"
```

**Production (.env.production):**
```bash
TOTP_ISSUER="Your App Name"
PROTO="https"
HOST="yourdomain.com"
# PORT not needed (no :port suffix for standard HTTPS)
```

**Important:**
- Never commit `.env` or `.env.production` files to version control
- Use strong, unique values for production
- Rotate credentials periodically

## Dependency Management

### Go Dependencies

```bash
# Update all Go modules
go get -u -t ./...
go mod tidy

# Audit for known vulnerabilities
go list -m all | nancy sleuth  # if using nancy
```

### Node.js Dependencies

```bash
# Update npm dependencies
cd ui
npm update

# Audit for vulnerabilities
npm audit
npm audit fix

# Check outdated packages
npm outdated
```

### Automated Scanning

This repository uses GitHub Dependabot for automated dependency vulnerability scanning:
- **Frequency:** Daily
- **Action:** Automatic pull requests for security updates
- **Review:** All Dependabot PRs should be reviewed and merged promptly

## Production Deployment Checklist

- [ ] Use HTTPS with valid SSL/TLS certificates
- [ ] Set strong, unique environment variables in `.env.production`
- [ ] Configure firewall rules (only expose necessary ports)
- [ ] Set up regular database backups
- [ ] Enable rate limiting on API endpoints
- [ ] Configure CORS with specific allowed origins
- [ ] Review and restrict file upload MIME types
- [ ] Set appropriate session token durations
- [ ] Enable PocketBase admin authentication
- [ ] Monitor logs for suspicious activity
- [ ] Keep all dependencies up to date
- [ ] Use a reverse proxy (nginx, Caddy) for TLS termination
- [ ] Consider using Redis for session storage (scaling)
- [ ] Implement monitoring and alerting

## Additional Resources

- [PocketBase Security Documentation](https://pocketbase.io/docs/)
- [WebAuthn Guide](https://webauthn.guide/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://go.dev/security/)

---

**Last Updated:** 2025-10-14

For questions about security, please contact the maintainers or open a discussion on GitHub.
