#!/usr/bin/env bash
# Status line de Claude Code: contexto, tokens de entrada/salida y costo estimado.
# Lee el JSON de la sesión por stdin e imprime una sola línea.
# Ej.: ctx 23% (46k/200k) · in 15.5k · out 1.2k · ~$0.012
#
# El costo (cost.total_cost_usd) es una estimación a precio de lista calculada
# por Claude Code, equivalente a pagar por API; no es la factura real.
# Los tokens son los del último response (no acumulados de la sesión).
set -euo pipefail

command -v jq >/dev/null 2>&1 || { echo "statusline: falta jq"; exit 0; }

input=$(cat)

# Formatea un entero de tokens: 950 -> 950, 15500 -> 15.5k, 1200000 -> 1.2M
fmt_tokens() {
  awk -v n="$1" 'BEGIN {
    if (n >= 1000000)      s = sprintf("%.1fM", n / 1000000)
    else if (n >= 1000)    s = sprintf("%.1fk", n / 1000)
    else                   s = sprintf("%d", n)
    if (s ~ /\.0[kM]$/) s = substr(s, 1, length(s) - 3) substr(s, length(s))
    printf "%s", s
  }'
}

# Un solo jq para extraer todo; "// 0" cubre campos null al inicio de sesión.
read -r pct window tin tout cost < <(
  jq -r '[
    (.context_window.used_percentage // 0),
    (.context_window.context_window_size // 0),
    (.context_window.total_input_tokens // 0),
    (.context_window.total_output_tokens // 0),
    (.cost.total_cost_usd // 0)
  ] | @tsv' <<<"$input"
)

pct_int=${pct%.*}
pct_int=${pct_int:-0}
used=$(( window * pct_int / 100 ))

# Color según uso del contexto: verde < 60, amarillo < 85, rojo después.
if   (( pct_int >= 85 )); then color=31
elif (( pct_int >= 60 )); then color=33
else                           color=32
fi

# Barra de progreso de BAR_WIDTH celdas con resolución de octavos de celda
# (▏▎▍▌▋▊▉█), así porcentajes cercanos se ven distintos. Lo no usado va en gris.
BAR_WIDTH=${BAR_WIDTH:-10}
eighths=$(awk -v p="$pct" -v w="$BAR_WIDTH" 'BEGIN {
  e = int(p * w * 8 / 100 + 0.5)
  if (e < 0) e = 0
  if (e > w * 8) e = w * 8
  print e
}')
full=$(( eighths / 8 ))
rem=$(( eighths % 8 ))
partials=("" "▏" "▎" "▍" "▌" "▋" "▊" "▉")

bar=""
for (( i = 0; i < full; i++ )); do bar+="█"; done
used_cells=$full
if (( rem > 0 )); then bar+="${partials[rem]}"; used_cells=$(( used_cells + 1 )); fi
empty=""
for (( i = used_cells; i < BAR_WIDTH; i++ )); do empty+="░"; done

pct_txt=$(printf '%.1f%%' "$pct")
tail_txt=$(printf ' (%s/%s) · in %s · out %s · ~$%.3f' \
  "$(fmt_tokens "$used")" "$(fmt_tokens "$window")" \
  "$(fmt_tokens "$tin")" "$(fmt_tokens "$tout")" "$cost")

# Alineación a la derecha. Claude Code exporta COLUMNS (tput no funciona aquí).
# Se mide el texto visible sin escapes de color; "·" ocupa 2 bytes pero 1 celda,
# por eso se sustituye por "." antes de contar.
# STATUSLINE_ALIGN=left desactiva el relleno; RIGHT_MARGIN reserva celdas a la
# derecha para el espaciado propio de la interfaz.
RIGHT_MARGIN=${RIGHT_MARGIN:-8}
pad=0
if [[ "${STATUSLINE_ALIGN:-right}" == "right" && "${COLUMNS:-0}" -gt 0 ]]; then
  plain=" ${pct_txt}${tail_txt}"
  ascii=${plain//·/.}
  visible=$(( BAR_WIDTH + ${#ascii} ))
  pad=$(( COLUMNS - visible - RIGHT_MARGIN ))
  (( pad < 0 )) && pad=0
fi

# El reset inicial evita que se recorten los espacios de relleno.
printf '\033[0m'
(( pad > 0 )) && printf '%*s' "$pad" ''
printf '\033[%sm%s\033[90m%s\033[0m' "$color" "$bar" "$empty"
printf ' \033[%sm%s\033[0m' "$color" "$pct_txt"
printf '%s\n' "$tail_txt"

# Segunda línea: gasto de la ventana de 5 horas y la semanal (solo suscripción).
# Si rate_limits no viene en el JSON, no se imprime nada.
now=$(date +%s)
fmt_reset() {
  local secs=$(( $1 - now )) d h m
  (( secs <= 0 )) && return 0
  d=$(( secs / 86400 )); h=$(( secs % 86400 / 3600 )); m=$(( secs % 3600 / 60 ))
  if   (( d > 0 )); then printf '%dd%dh' "$d" "$h"
  elif (( h > 0 )); then printf '%dh%02dm' "$h" "$m"
  else                   printf '%dm' "$m"
  fi
}

LIMIT_BAR_WIDTH=${LIMIT_BAR_WIDTH:-8}
limit_colored=""   # texto con colores
limit_plain=""     # texto visible, para medir el relleno
for entry in "five_hour:5h" "seven_day:sem"; do
  key=${entry%%:*}; label=${entry##*:}
  read -r lpct lreset < <(
    jq -r --arg k "$key" '
      (.rate_limits[$k] // empty)
      | [ (.used_percentage // .percentUsed // 0),
          ((.resets_at // .resetsAt // 0)
            | if type == "number" then . else (try fromdateiso8601 catch 0) end) ]
      | @tsv' <<<"$input"
  ) || true
  [[ -z "${lpct:-}" ]] && continue
  lint=${lpct%.*}; lint=${lint:-0}
  if   (( lint >= 85 )); then lc=31
  elif (( lint >= 60 )); then lc=33
  else                        lc=32
  fi
  lfull=$(awk -v p="$lpct" -v w="$LIMIT_BAR_WIDTH" 'BEGIN { e = int(p * w * 8 / 100 + 0.5); if (e > w * 8) e = w * 8; print int(e / 8) }')
  lrem=$(awk -v p="$lpct" -v w="$LIMIT_BAR_WIDTH" 'BEGIN { e = int(p * w * 8 / 100 + 0.5); if (e > w * 8) e = w * 8; print e % 8 }')
  lbar=""; for (( i = 0; i < lfull; i++ )); do lbar+="█"; done
  lcells=$lfull
  if (( lrem > 0 )); then lbar+="${partials[lrem]}"; lcells=$(( lcells + 1 )); fi
  lempty=""; for (( i = lcells; i < LIMIT_BAR_WIDTH; i++ )); do lempty+="░"; done
  ltxt=$(printf '%.1f%%' "$lpct")
  lreset_txt=$(fmt_reset "${lreset:-0}")
  lrest=""; [[ -n "$lreset_txt" ]] && lrest=" (reset ${lreset_txt})"
  if [[ -n "$limit_plain" ]]; then
    limit_plain+="  ·  "; limit_colored+="  ·  "
  fi
  limit_plain+="${label} ${lbar}${lempty} ${ltxt}${lrest}"
  limit_colored+="${label} \033[${lc}m${lbar}\033[90m${lempty}\033[0m \033[${lc}m${ltxt}\033[0m\033[90m${lrest}\033[0m"
done

if [[ -n "$limit_plain" ]]; then
  lpad=0
  if [[ "${STATUSLINE_ALIGN:-right}" == "right" && "${COLUMNS:-0}" -gt 0 ]]; then
    lascii=${limit_plain//·/.}
    lpad=$(( COLUMNS - ${#lascii} - RIGHT_MARGIN ))
    (( lpad < 0 )) && lpad=0
  fi
  printf '\033[0m'
  (( lpad > 0 )) && printf '%*s' "$lpad" ''
  printf '%b\n' "$limit_colored"
fi
