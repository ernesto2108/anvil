# Guía de Patrones de Seguridad

## Prevención de SQL Injection

El patrón de seguridad más crítico. Siempre usa consultas parametrizadas.

```go
// VULNERABLE — input del usuario interpolado en el SQL
email := r.FormValue("email")
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
// Ataque: email = "'; DROP TABLE users; --"
// Resultado: SELECT * FROM users WHERE email = ''; DROP TABLE users; --'

// SEGURO — consulta parametrizada, el driver escapa automáticamente
query := "SELECT * FROM users WHERE email = $1"
row := db.QueryRowContext(ctx, query, email)
// El driver trata el email como DATO, nunca como código SQL

// SEGURO — con query builder (patrón strings.Builder)
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

**Reglas:**
- NUNCA uses `fmt.Sprintf` o concatenación de strings para SQL con input de usuario
- Siempre usa placeholders `$1, $2, $N` (PostgreSQL) o `?` (MySQL)
- Las funciones de query builder deben retornar `(string, []any, error)` — ver patrones de base de datos en SKILL.md
- Usa `sqlc` o generadores de código similares para validación de SQL en tiempo de compilación cuando sea posible

## Aleatoriedad Criptográfica

```go
// INSEGURO — math/rand es determinístico, predecible
import "math/rand"
token := fmt.Sprintf("%d", rand.Int63())
// Un atacante puede predecir la secuencia si conoce el seed

// SEGURO — crypto/rand usa entropía del SO, impredecible
import "crypto/rand"

// Go 1.24+
token := rand.Text()

// Go < 1.24
b := make([]byte, 32)
if _, err := crypto_rand.Read(b); err != nil {
    return fmt.Errorf("generate token: %w", err)
}
token := hex.EncodeToString(b) // string hex de 64 chars
```

**Usa crypto/rand para:** tokens de sesión, API keys, tokens de reset de contraseña, tokens CSRF, nonces, cualquier valor que un atacante se beneficiaría de predecir.

## Validación de Input

```go
// MALO — blacklist (siempre te olvidas de algo)
if strings.Contains(input, "<script>") {
    return errors.New("invalid input")
}
// Evadido con: <SCRIPT>, <img onerror=...>, %3Cscript%3E, etc.

// BUENO — whitelist (solo permite valores conocidos como válidos)
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

// BUENO — valida formato con regex para campos de forma libre
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
    if !emailRegex.MatchString(email) {
        return fmt.Errorf("invalid email format: %q", email)
    }
    return nil
}

// BUENO — valida rangos numéricos
func ValidateAmount(amount int64) error {
    if amount < 0 || amount > 10_000_000 { // máximo 100,000.00 en centavos
        return fmt.Errorf("amount %d out of range", amount)
    }
    return nil
}
```

**Reglas:**
- Valida en los límites del sistema (handlers HTTP, consumers de mensajes) — confía en el código interno
- Whitelist > blacklist — define qué está permitido, rechaza todo lo demás
- Valida tipo, formato, rango y longitud
- Retorna mensajes de error claros (qué estaba mal, qué se esperaba)

## Límites de Tamaño del Body de la Request

```go
// SIN límite — el atacante envía 10GB, el servidor se queda sin memoria
body, _ := io.ReadAll(r.Body) // OOM potencial → crash del servidor (DoS)

// CON límite — máximo 1MB, retorna error si se supera
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
var input CreateUserRequest
if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
    // MaxBytesReader retorna un error específico si el body supera el límite
    http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
    return
}
```

**Aplica en middleware para límite global, o por handler para límites distintos** (subida de archivos = 10MB, JSON API = 1MB).

## Security Headers

```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Previene sniffing del tipo MIME
        w.Header().Set("X-Content-Type-Options", "nosniff")
        // Previene clickjacking vía iframes
        w.Header().Set("X-Frame-Options", "DENY")
        // Solo permite recursos del mismo origen
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        // Fuerza HTTPS
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

        next.ServeHTTP(w, r)
    })
}
```

## Escaneo de Vulnerabilidades en Dependencias

```bash
# Instala govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Escanea todos los packages contra la base de datos de vulnerabilidades de Go
govulncheck ./...

# Ejecuta en CI — falla el build ante vulnerabilidades conocidas
# Agrega a GitHub Actions o pre-commit hook
```

**Reglas:**
- Ejecuta `govulncheck` en CI en cada PR
- Actualiza las dependencias vulnerables de inmediato para CVEs críticos/altos
- Audita los cambios en `go.sum` en los code reviews — nuevas dependencias = nueva superficie de ataque

## Gestión de Secretos

```go
// NUNCA hardcodees secretos
const apiKey = "sk-1234567890" // NO — queda en el historial de git para siempre

// NUNCA registres secretos en logs
logger.Info("auth", slog.String("token", token)) // NO — aparece en los agregadores de logs

// Carga desde el entorno, valida al arrancar
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    log.Fatal("API_KEY environment variable is required")
}
```

**Reglas:**
- Secretos vía variables de entorno o secret managers (AWS SSM, Vault, GCP Secret Manager)
- Nunca commités archivos `.env` con secretos reales — usa `.env.example` con placeholders
- Rota secretos periódicamente — diseña para rotación (sin expiración hardcodeada)
- Limita el alcance de los secretos — cada servicio obtiene solo los secretos que necesita
