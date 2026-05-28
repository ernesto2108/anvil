---
# NOTE: This file is documentation/index only — NOT an invocable slash command.
# It intentionally omits the `name:` field so the CLI does not register it.
description: Escritura y revisión de mensajes de commit Git siguiendo la spec de Conventional Commits
invocable: false
---

# Skills de Git Commit

Comandos slash para escribir y revisar mensajes de commit Git siguiendo las mejores prácticas de la industria.

## Comandos

### `/git:commit`
Analiza los cambios staged (`git diff --cached`) y escribe un mensaje de commit convencional. Pide confirmación antes de hacer el commit.

**Uso:**
```
/git:commit
```

**Qué hace:**
1. Lee el diff staged y el nombre de la rama
2. Determina el tipo y scope apropiado del commit
3. Escribe un mensaje de commit conforme a la spec
4. Pide confirmación antes de hacer el commit

### `/git:commit-review`
Revisa los últimos N commits (por defecto 5) y puntúa cada uno contra la spec de Conventional Commits.

**Uso:**
```
/git:commit-review        # revisa los últimos 5 commits
/git:commit-review 10     # revisa los últimos 10 commits
```

**Qué hace:**
1. Lee los mensajes de commits recientes
2. Puntúa cada uno en 12 criterios (estructura, contenido, mejores prácticas)
3. Reporta problemas y sugiere reescrituras para commits que fallan
4. Proporciona un resumen con problemas comunes

**Escala de puntuación:**
- 90-100: Excelente
- 70-89: Aceptable
- 50-69: Necesita mejora
- 0-49: Calidad pobre

### `/git:message`
Genera un mensaje de commit a partir de una descripción en lenguaje natural sin hacer el commit.

**Uso:**
```
/git:message added login with Google OAuth
/git:message fixed crash on empty input in parser
/git:message renamed endpoints to follow REST conventions, breaks clients
```

**Qué hace:**
1. Parsea el lenguaje natural en tipo, scope y descripción
2. Genera un mensaje de commit convencional correctamente formateado
3. Muestra el mensaje — NO hace el commit

## Estándares aplicados

Basado en:
- [Conventional Commits v1.0.0](https://www.conventionalcommits.org/en/v1.0.0/)
- [How to Write a Git Commit Message](https://cbea.ms/git-commit/) (Chris Beams)
- [Angular Commit Convention](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)
- [Semantic Release](https://semantic-release.gitbook.io/semantic-release/)

### Tipos de commit

| Tipo | Propósito | Impacto SemVer |
|------|-----------|----------------|
| `feat` | Nueva funcionalidad | MINOR |
| `fix` | Corrección de bug | PATCH |
| `docs` | Documentación | — |
| `style` | Formateo | — |
| `refactor` | Reestructuración | — |
| `test` | Tests | — |
| `chore` | Mantenimiento | — |
| `perf` | Rendimiento | — |
| `ci` | CI/CD | — |
| `build` | Sistema de build | — |

### Reglas

1. Línea de asunto máx 50 caracteres
2. Modo imperativo ("add" no "added")
3. Sin punto al final del asunto
4. Línea en blanco entre asunto y cuerpo
5. Cuerpo se ajusta a 72 caracteres
6. El cuerpo explica QUÉ y POR QUÉ, no CÓMO
7. Breaking changes usan sufijo `!` Y footer `BREAKING CHANGE:`
8. Referencias a issues en el footer (`Closes #123`)
9. Sin mensajes vagos ("fix bug", "update", "WIP")

## Alcance

Estos comandos se enfocan exclusivamente en la calidad del **mensaje** de commit. La revisión de código la maneja CodeRabbit.
