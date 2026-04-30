#!/usr/bin/env bash
# verify-handoff.sh — Gate ejecutable que valida la integridad de un handoff
# antes de que el orquestador avance al siguiente agente.
#
# Uso:
#   verify-handoff.sh <PROJECT_ROOT> <TASK_ID>
#
# Ejemplo:
#   verify-handoff.sh /Users/ernesto/projects/blt-bookings BLT-123
#
# Exit codes:
#   0  — handoff válido, el orquestador puede avanzar
#   1  — handoff con gaps, no avanzar (mensaje en stderr)
#   2  — error de uso (args inválidos)
#
# El script es deliberadamente simple (bash + grep) para que corra en cualquier
# proyecto sin compilar nada. Las reglas que valida coinciden con
# skills/handoff/template.md.

set -euo pipefail

# ───────── args ─────────
if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <PROJECT_ROOT> <TASK_ID>" >&2
  exit 2
fi

PROJECT_ROOT="$1"
TASK_ID="$2"
HANDOFF_PATH="${PROJECT_ROOT}/.handoff/${TASK_ID}.md"

# ───────── helpers ─────────
fail() {
  echo "BLOCKED: $1" >&2
  exit 1
}

require_section() {
  local section="$1"
  grep -qE "^${section}\b" "$HANDOFF_PATH" || fail "missing section '${section}' in ${HANDOFF_PATH}"
}

# ───────── validaciones ─────────

# 1. El archivo existe
[ -f "$HANDOFF_PATH" ] || fail "handoff file not found at ${HANDOFF_PATH}"

# 2. Tiene contenido real (> 50 líneas — descarta plantillas vacías)
LINES=$(wc -l < "$HANDOFF_PATH" | tr -d ' ')
if [ "$LINES" -lt 50 ]; then
  fail "handoff too short (${LINES} lines, min 50). Likely empty/template."
fi

# 3. Secciones obligatorias presentes
require_section "## Input recibido"
require_section "## Archivos modificados"
require_section "## Decisiones tomadas"
require_section "## Handoff for tester"
require_section "## Output entregado"

# 4. Tiene "## Estado actual" o "## Fases" (cross-stack usa Fases)
if ! grep -qE '^## (Estado actual|Fases)\b' "$HANDOFF_PATH"; then
  fail "missing '## Estado actual' or '## Fases' (cross-stack) in ${HANDOFF_PATH}"
fi

# 5. La tabla de "Output entregado" tiene PASS en Build y Lint
#    Extraemos líneas entre "## Output entregado" y la siguiente sección "## ".
#    El matcher es flexible: acepta nombres de fila como "Build", "Build (Go)",
#    "go build ./...", "npm run build" — cualquier fila que contenga "build" /
#    "lint" en su primera columna.
OUTPUT_BLOCK=$(awk '/^## Output entregado/{flag=1; next} /^## /{flag=0} flag' "$HANDOFF_PATH" || true)

if [ -z "$OUTPUT_BLOCK" ]; then
  fail "'## Output entregado' section is empty"
fi

# Verifica que existe al menos una fila de tabla con "build" en la primera col
# y que su última col contiene PASS / OK / 0
BUILD_ROW=$(echo "$OUTPUT_BLOCK" | grep -iE '^\s*\|[^|]*build[^|]*\|' | grep -viE '^\s*\|\s*-+' || true)
if [ -z "$BUILD_ROW" ]; then
  fail "'## Output entregado' has no Build row (expected a table row mentioning 'build')"
fi
if ! echo "$BUILD_ROW" | grep -iE '\|\s*(PASS|OK|✅|0)\s*\|?\s*$' >/dev/null; then
  fail "Build row in '## Output entregado' is not PASS/OK. Developer must fix build before tester."
fi

# Lint: misma lógica. Aceptamos PASS / OK / "0 issues" / "0 new issues" / "0".
LINT_ROW=$(echo "$OUTPUT_BLOCK" | grep -iE '^\s*\|[^|]*(lint|clippy|eslint|ruff|golangci|dart analyze)[^|]*\|' | grep -viE '^\s*\|\s*-+' || true)
if [ -z "$LINT_ROW" ]; then
  fail "'## Output entregado' has no Lint row (expected a table row mentioning 'lint' or known linter name)"
fi
if ! echo "$LINT_ROW" | grep -iE '\|\s*(PASS|OK|✅|0( (new )?issues)?)\s*\|?\s*$' >/dev/null; then
  fail "Lint row in '## Output entregado' is not PASS / 0 issues. Developer must fix lint before tester."
fi

# 6. Tests requeridos: al menos 1 test enumerado en "Handoff for tester"
TESTER_BLOCK=$(awk '/^## Handoff for tester/,/^## [^H]/' "$HANDOFF_PATH" || true)

# Buscar líneas tipo "1. nombre_del_test" o "  1. nombre" dentro del bloque del tester
if ! echo "$TESTER_BLOCK" | grep -qE '^\s*[0-9]+\.\s+\S' ; then
  fail "'## Handoff for tester' has no enumerated tests (expected lines like '1. test_name')"
fi

# ───────── éxito ─────────
echo "OK: ${HANDOFF_PATH} (${LINES} lines, all required sections present, Build/Lint PASS)"
exit 0
