# Riesgos y Deuda Técnica — <ProjectName>

last_updated: <YYYY-MM-DD>

## Gotchas operativos

<!-- Comportamientos que sorprenden, restricciones de runtime, edge cases conocidos -->

### <Nombre del gotcha>
- **Dónde:** `<path>:<line>`
- **Descripción:** <qué pasa y por qué sorprende>
- **Workaround:** <cómo manejarlo>

## Deuda técnica

<!-- Archivos > 300 líneas, TODOs con impacto, acoplamiento conocido -->

### Archivos candidatos a refactor

| Archivo | Líneas | Razón |
|---------|--------|-------|
| `<path>` | XXX | <violación SRP / demasiada responsabilidad> |

### TODOs y FIXMEs con impacto

```bash
# Detectar con:
grep -rn "TODO\|FIXME\|HACK\|XXX" --include="*.go" --include="*.ts" | grep -v "_test"
```

| Ubicación | Mensaje | Impacto estimado |
|-----------|---------|-----------------|
| `<path>:<line>` | `<mensaje>` | <bajo/medio/alto> |

## Restricciones conocidas

<!-- Limitaciones de la plataforma, dependencias externas, decisiones de terceros -->

- **<Restricción>:** <descripción y qué afecta>

## Dependencias frágiles

<!-- Servicios externos sin retry, dependencias de versión fija, CGO, etc. -->

- **<Dep>:** <por qué es frágil y mitigación>

## Áreas sin tests

<!-- Dominios o funciones críticas sin cobertura -->

- `<path>` — <por qué importa tener tests aquí>
