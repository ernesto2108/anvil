# Estructura de Archivos de Test y Tests con Tabla

## Estructura de Archivos de Test

Dos estilos comunes dependiendo de las convenciones del proyecto:

### solo stdlib (sin dependencias externas de test)

```go
package user_test  // use external test package for black-box testing

import (
    "context"
    "testing"

    "myapp/internal/user"
)

// Test helpers at the top
func newTestUser(t *testing.T, email string) *user.User {
    t.Helper()
    u, err := user.New(email, "Test User")
    if err != nil {
        t.Fatalf("newTestUser: %v", err)
    }
    return u
}

// Tests grouped by function/method
func Test_New(t *testing.T) { ... }
func Test_User_Activate(t *testing.T) { ... }
func Test_User_ChangeEmail(t *testing.T) { ... }
```

### testify (proyectos que usan stretchr/testify)

```go
package user_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "myapp/internal/user"
)

func newTestUser(t *testing.T, email string) *user.User {
    t.Helper()
    u, err := user.New(email, "Test User")
    require.NoError(t, err)
    return u
}

func Test_New(t *testing.T) { ... }
func Test_User_Activate(t *testing.T) { ... }
```

- Usar sufijo `_test` en el paquete para tests de caja negra (testean la API pública)
- Usar el mismo paquete para tests de caja blanca solo cuando se testea lógica no exportada
- Nombrar tests: `Test_FunctionName` o `Test_Type_Method` (guion bajo después de Test)
- Elegir un estilo de assertion por proyecto y ser consistente

---

## Tests con Tabla (Table-Driven Tests)

El patrón por defecto para cualquier función con múltiples escenarios de entrada:

```go
func Test_ParseAmount(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int64
        wantErr string // empty = no error expected
    }{
        {name: "valid dollars and cents", input: "12.50", want: 1250},
        {name: "whole number", input: "100", want: 10000},
        {name: "zero", input: "0", want: 0},
        {name: "negative", input: "-5.00", want: -500},
        {name: "too many decimals", input: "1.234", wantErr: "invalid amount"},
        {name: "not a number", input: "abc", wantErr: "invalid amount"},
        {name: "empty string", input: "", wantErr: "empty input"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseAmount(tt.input)

            if tt.wantErr != "" {
                // stdlib: use t.Fatalf/t.Errorf
                if err == nil {
                    t.Fatalf("expected error containing %q, got nil", tt.wantErr)
                }
                if !strings.Contains(err.Error(), tt.wantErr) {
                    t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
                }
                return

                // testify alternative:
                // require.Error(t, err)
                // assert.Contains(t, err.Error(), tt.wantErr)
            }

            // stdlib
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.want {
                t.Errorf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
            }

            // testify alternative:
            // require.NoError(t, err)
            // assert.Equal(t, tt.want, got)
        })
    }
}
```

Reglas:
- Cada caso de test tiene un `name` descriptivo
- stdlib: usar `t.Fatalf` para verificaciones fatales (detiene en fallo), `t.Errorf` para no-fatales
- testify: usar `require` para verificaciones fatales, `assert` para no-fatales
- Los casos de error verifican el mensaje/tipo del error, no solo `err != nil`
- Mantener los datos de test inline cuando son simples, usar `testdata/` para fixtures complejos
