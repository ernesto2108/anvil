# Security Patterns Guide

## SQL Injection Prevention

The most critical security pattern. Always use parameterized queries.

```go
// VULNERABLE — user input interpolated into SQL
email := r.FormValue("email")
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
// Attack: email = "'; DROP TABLE users; --"
// Result: SELECT * FROM users WHERE email = ''; DROP TABLE users; --'

// SECURE — parameterized query, driver escapes automatically
query := "SELECT * FROM users WHERE email = $1"
row := db.QueryRowContext(ctx, query, email)
// The driver treats email as DATA, never as SQL code

// SECURE — with query builder (strings.Builder pattern)
var b strings.Builder
var args []any
b.WriteString("SELECT id, name FROM users WHERE 1=1")
if email != "" {
    args = append(args, email)
    fmt.Fprintf(&b, " AND email = $%d", len(args))
}
if status != "" {
    args = append(args, status)
    fmt.Fprintf(&b, " AND status = $%d", len(args))
}
rows, err := db.QueryContext(ctx, b.String(), args...)
```

**Rules:**
- NEVER use `fmt.Sprintf` or string concatenation for SQL with user input
- Always use `$1, $2, $N` placeholders (PostgreSQL) or `?` (MySQL)
- Query builder functions should return `(string, []any, error)` — see database patterns in SKILL.md
- Use `sqlc` or similar code generators for compile-time SQL validation when possible

## Cryptographic Randomness

```go
// INSECURE — math/rand is deterministic, predictable
import "math/rand"
token := fmt.Sprintf("%d", rand.Int63())
// An attacker can predict the sequence if they know the seed

// SECURE — crypto/rand uses OS entropy, unpredictable
import "crypto/rand"

// Go 1.24+
token := rand.Text()

// Go < 1.24
b := make([]byte, 32)
if _, err := crypto_rand.Read(b); err != nil {
    return fmt.Errorf("generate token: %w", err)
}
token := hex.EncodeToString(b) // 64-char hex string
```

**Use crypto/rand for:** session tokens, API keys, password reset tokens, CSRF tokens, nonces, any value an attacker would benefit from predicting.

## Input Validation

```go
// BAD — blacklist (you always miss something)
if strings.Contains(input, "<script>") {
    return errors.New("invalid input")
}
// Bypassed with: <SCRIPT>, <img onerror=...>, %3Cscript%3E, etc.

// GOOD — whitelist (only allow known-valid values)
func ValidateStatus(s string) error {
    valid := map[string]bool{
        "active":   true,
        "inactive": true,
        "pending":  true,
    }
    if !valid[s] {
        return fmt.Errorf("invalid status: %q", s)
    }
    return nil
}

// GOOD — validate format with regex for free-form fields
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
    if !emailRegex.MatchString(email) {
        return fmt.Errorf("invalid email format: %q", email)
    }
    return nil
}

// GOOD — validate numeric ranges
func ValidateAmount(amount int64) error {
    if amount < 0 || amount > 10_000_000 { // max 100,000.00 in cents
        return fmt.Errorf("amount %d out of range", amount)
    }
    return nil
}
```

**Rules:**
- Validate at system boundaries (HTTP handlers, message consumers) — trust internal code
- Whitelist > blacklist — define what's allowed, reject everything else
- Validate type, format, range, and length
- Return clear error messages (what was wrong, what was expected)

## Request Body Limits

```go
// WITHOUT limit — attacker sends 10GB, server runs out of memory
body, _ := io.ReadAll(r.Body) // potential OOM → server crash (DoS)

// WITH limit — max 1MB, returns error if exceeded
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
var input CreateUserRequest
if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
    // MaxBytesReader returns specific error if body exceeds limit
    http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
    return
}
```

**Apply in middleware for global limit, or per-handler for different limits** (file upload = 10MB, JSON API = 1MB).

## Security Headers

```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent MIME type sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")
        // Prevent clickjacking via iframes
        w.Header().Set("X-Frame-Options", "DENY")
        // Only allow resources from same origin
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        // Force HTTPS
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

        next.ServeHTTP(w, r)
    })
}
```

## Dependency Vulnerability Scanning

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan all packages against Go vulnerability database
govulncheck ./...

# Run in CI — fail build on known vulnerabilities
# Add to GitHub Actions or pre-commit hook
```

**Rules:**
- Run `govulncheck` in CI on every PR
- Update vulnerable dependencies immediately for critical/high CVEs
- Audit `go.sum` changes in code reviews — new dependencies = new attack surface

## Secrets Management

```go
// NEVER hardcode secrets
const apiKey = "sk-1234567890" // NO — ends up in git history forever

// NEVER log secrets
logger.Info("auth", slog.String("token", token)) // NO — appears in log aggregators

// Load from environment, validate at startup
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    log.Fatal("API_KEY environment variable is required")
}
```

**Rules:**
- Secrets via environment variables or secret managers (AWS SSM, Vault, GCP Secret Manager)
- Never commit `.env` files with real secrets — use `.env.example` with placeholders
- Rotate secrets periodically — design for rotation (no hardcoded expiration)
- Limit secret scope — each service gets only the secrets it needs
