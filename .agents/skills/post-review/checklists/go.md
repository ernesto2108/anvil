# Checklist de Revisión Go

## Correctness

- [ ] Todos los `error` retornados se verifican (`if err != nil`)
- [ ] No hay goroutines sin mecanismo de cancelacion (context, done channel)
- [ ] No hay race conditions en acceso concurrente a maps o slices
- [ ] Defer se usa correctamente (no dentro de loops sin intencion)
- [ ] Los channels se cierran por el productor, no el consumidor
- [ ] No hay nil pointer dereference en structs retornados por funciones que pueden fallar
- [ ] Los type assertions incluyen el ok check (`val, ok := x.(Type)`)
- [ ] Las iteraciones sobre maps no asumen orden

## Security

- [ ] No hay SQL concatenado — usar queries parametrizadas (`$1`, `$2`)
- [ ] No hay secrets hardcodeados (API keys, passwords, tokens)
- [ ] Input del usuario se valida antes de usar
- [ ] No hay `fmt.Sprintf` para construir SQL, HTML o comandos shell
- [ ] Los endpoints autenticados verifican permisos
- [ ] No se loguean datos sensibles (passwords, tokens, PII)
- [ ] Los headers CORS estan configurados correctamente

## Conventions

- [ ] Nombres de funciones exportadas en PascalCase
- [ ] Nombres de funciones privadas en camelCase
- [ ] Interfaces pequenas (1-3 metodos)
- [ ] Errores custom implementan `error` interface
- [ ] Paquetes nombrados en singular, lowercase, sin underscores
- [ ] No hay `init()` functions sin justificacion clara
- [ ] Structs con campos ordenados por alineacion de memoria

## Tests

- [ ] Funciones nuevas tienen test correspondiente
- [ ] Tests usan table-driven pattern cuando hay multiples casos
- [ ] Tests cubren el happy path Y al menos un error path
- [ ] No hay tests que dependen de estado externo (DB, API, filesystem) sin cleanup
- [ ] Mocks implementan la interface, no son structs genericos

## Performance

- [ ] No hay queries dentro de loops (N+1)
- [ ] Slices pre-allocados con `make([]T, 0, expectedCap)` cuando el size es conocido
- [ ] No hay string concatenation en loops (usar `strings.Builder`)
- [ ] Context se propaga correctamente en cadenas de llamadas
- [ ] No hay `reflect` en hot paths
- [ ] Conexiones a DB/HTTP clients se reusan, no se crean por request
