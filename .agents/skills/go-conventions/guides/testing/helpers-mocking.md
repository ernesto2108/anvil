# Helpers de Test y Mocking con Mockery

## Helpers de Test

```go
// Siempre marcar los helpers con t.Helper() — los errores reportan la línea del caller
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    db, err := sql.Open("postgres", testDSN)
    if err != nil {
        t.Fatalf("open test db: %v", err)
    }

    t.Cleanup(func() {
        db.Close()
    })

    return db
}

// Factory helpers para objetos de test comunes
func newTestOrder(t *testing.T, opts ...func(*Order)) *Order {
    t.Helper()
    o := &Order{
        ID:     uuid.New(),
        Status: StatusPending,
        Amount: 1000,
    }
    for _, opt := range opts {
        opt(o)
    }
    return o
}
```

- `t.Helper()` en cada helper — hace que el output de errores apunte al test, no al helper
- `t.Cleanup()` en lugar de `defer` — ejecuta después del test sin importar el anidamiento de subtests
- Las factory functions usan functional options para mayor flexibilidad

---

## Mocking con Mockery (OBLIGATORIO)

**Nunca escribir mocks manuales.** Los mocks manuales pueden divergir de la interfaz real — el test pasa verde pero el código está roto. Mockery genera mocks desde las interfaces Go, garantizando que siempre coincidan.

### Prerequisito: instalar mockery

```bash
brew install mockery    # macOS
# o descargar el binario desde https://github.com/vektra/mockery/releases
```

### Configuración del proyecto

Crear `.mockery.yaml` en la raíz del módulo Go (si no existe):

```yaml
packages:
  github.com/tu/modulo:
    config:
      dir: "{{.InterfaceDir}}"
      filename: "mock_{{.InterfaceName | snakecase}}_test.go"
      pkgname: "{{.SrcPackageName}}_test"
      inpackage: true
      with-expecter: true
```

Campos clave:
- `dir: "{{.InterfaceDir}}"` — genera el mock junto al código fuente
- `pkgname: "..._test"` — el mock vive en el paquete de test (caja negra)
- `with-expecter: true` — habilita la API `.EXPECT()` type-safe
- `inpackage: true` — genera en el mismo directorio, no en carpeta `mocks/`

### Generar mocks

```bash
mockery    # lee .mockery.yaml y genera todos los mocks configurados
```

Ejecutar cada vez que una interfaz cambie. Si el mock no compila, la interfaz cambió y el test lo detecta inmediatamente.

### Usar mocks en tests

```go
func Test_CreateUser_success(t *testing.T) {
    mockRepo := NewMockUserRepository(t)  // auto-registra cleanup y AssertExpectations

    mockRepo.EXPECT().
        Save(mock.Anything, mock.MatchedBy(func(u *User) bool {
            return u.Email == "test@example.com"
        })).
        Return(nil).
        Once()

    svc := NewUserService(mockRepo)
    err := svc.Create(context.Background(), "test@example.com", "Test")
    require.NoError(t, err)
}

func Test_CreateUser_repoError(t *testing.T) {
    mockRepo := NewMockUserRepository(t)

    mockRepo.EXPECT().
        Save(mock.Anything, mock.Anything).
        Return(errors.New("db connection lost")).
        Once()

    svc := NewUserService(mockRepo)
    err := svc.Create(context.Background(), "test@example.com", "Test")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "db connection lost")
}
```

### Patrones clave

| Patrón | Cuándo usar |
|---|---|
| `mock.Anything` | No te importa el valor exacto del argumento |
| `mock.MatchedBy(fn)` | Verificar propiedades específicas del argumento |
| `.Return(val)` | Configurar valor de retorno |
| `.Once()` / `.Times(n)` | Verificar que se llamó exactamente N veces |
| `.Maybe()` | La llamada puede o no ocurrir (no falla si no se llama) |
| `NewMock*(t)` | Pasar `t` al constructor — auto-llama `AssertExpectations` en cleanup |

### Reglas

1. **Siempre pasar `t` al constructor del mock** — esto registra `AssertExpectations` automáticamente via `t.Cleanup()`
2. **Usar `.EXPECT()`** (con expecter) en vez de `.On()` — es type-safe y el compilador detecta errores de firma
3. **Un mock por interfaz** — no mockear structs concretos, solo interfaces
4. **Regenerar después de cambiar interfaces** — ejecutar `mockery` antes de correr tests si tocaste una interfaz
5. **Si mockery no está instalado o falla** — NO recurrir a mocks manuales. Reportar al orquestador: "Mockery no disponible — necesito que se instale antes de continuar." El developer o devops debe resolver la instalación
6. **Prohibido sqlmock/httpmock** — para DB testear contra SQLite real; para HTTP usar `httptest.NewServer` de la stdlib
