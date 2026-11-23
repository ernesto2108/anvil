// anvil-usage: dibuja el uso de contexto y el costo estimado a la derecha de la
// linea de modo (SessionMode, donde sale "auto mode on (shift+tab to cycle)").
//
// "in" y "out" son los tokens de la ultima respuesta del modelo
// (context.apiUsage): "in" suma input fresco y cache (escritura y lectura).
// Debajo va una fila con el gasto de la ventana de 5 horas y la semanal.
// Si un arbol no valida, Claude Code dibuja su propio SessionMode y deja una
// linea en el transcript con el motivo.

const BAR_WIDTH = 10
const PARTIALS = ['', '▏', '▎', '▍', '▌', '▋', '▊', '▉']
const REFRESH_MS = 2000

// 950 -> "950", 15500 -> "15.5k", 1000000 -> "1M"
function fmtTokens(n) {
  const v = Number(n) || 0
  const s = v >= 1e6 ? (v / 1e6).toFixed(1) + 'M' : v >= 1e3 ? (v / 1e3).toFixed(1) + 'k' : String(Math.round(v))
  return s.replace(/\.0([kM])$/, '$1')
}

// Barra con resolucion de octavos de celda: [lleno, vacio]
function bar(percent, width = BAR_WIDTH) {
  const clamped = Math.min(100, Math.max(0, Number(percent) || 0))
  const eighths = Math.round((clamped * width * 8) / 100)
  const full = Math.floor(eighths / 8)
  const rem = eighths % 8
  const filled = '█'.repeat(full) + PARTIALS[rem]
  const empty = '░'.repeat(width - full - (rem > 0 ? 1 : 0))
  return [filled, empty]
}

// Verde < 60, amarillo < 85, rojo despues
function colorFor(percent) {
  return percent >= 85 ? 'red' : percent >= 60 ? 'yellow' : 'green'
}

// La forma de usage().cost no esta confirmada: acepta un numero o un objeto
// con alguno de los nombres habituales. Devuelve USD o null.
function costUsd(cost) {
  if (typeof cost === 'number') return cost
  if (cost && typeof cost === 'object') {
    for (const key of ['usd', 'totalUsd', 'total_cost_usd', 'total']) {
      if (typeof cost[key] === 'number') return cost[key]
    }
  }
  return null
}

// "2h10m", "3d4h", "35m" hasta el reset; '' si no hay fecha valida
function fmtReset(iso) {
  const ms = Date.parse(iso) - Date.now()
  if (!Number.isFinite(ms) || ms <= 0) return ''
  const m = Math.floor(ms / 60000)
  const d = Math.floor(m / 1440)
  const h = Math.floor((m % 1440) / 60)
  if (d > 0) return d + 'd' + h + 'h'
  if (h > 0) return h + 'h' + String(m % 60).padStart(2, '0') + 'm'
  return m + 'm'
}

const WINDOWS = [
  ['five_hour', '5h'],
  ['seven_day', 'sem'],
]

// Fila con el gasto de la ventana de 5 horas y la semanal (solo suscripcion)
function limitsRow(Box, Text, rateLimits) {
  const parts = []
  for (const [kind, label] of WINDOWS) {
    const w = (rateLimits || []).find((r) => r.kind === kind)
    if (!w) continue
    const pct = Number(w.percentUsed) || 0
    const color = colorFor(pct)
    const [filled, empty] = bar(pct, 8)
    const reset = fmtReset(w.resetsAt)
    if (parts.length) parts.push(Text({ dimColor: true, children: ['  ·  '] }))
    parts.push(
      Text({ children: [label + ' '] }),
      Text({ color, children: [filled] }),
      Text({ dimColor: true, children: [empty] }),
      Text({ color, children: [' ' + pct.toFixed(1) + '%'] }),
      Text({ dimColor: true, children: [reset ? ' (reset ' + reset + ')' : ''] }),
    )
  }
  if (!parts.length) return null
  return Box({ flexDirection: 'row', justifyContent: 'flex-end', width: '100%', children: parts })
}

export function register(on) {
  // Refresca el dibujo en cada medicion (tras cada turno) y por timer, porque
  // Claude Code no redibuja solo cuando cambian los datos del mod.
  on('session.start', async ($, e, next) => {
    $.clock.every(REFRESH_MS, () => $.ui.invalidate('ui.render'))
    return next(e)
  })

  on('session.measure', async ($, e, next) => {
    $.ui.invalidate('ui.render')
    return next(e)
  })

  on('ui.render', { component: 'SessionMode' }, async ($, e, next) => {
    const { Box, Text } = $.ui.resolve(e)
    // El dibujo propio de Claude Code (auto mode on ...), que se conserva
    const theirs = await next(e)

    const usage = await $.session.usage()
    const percent = Number(usage?.context?.percent) || 0
    const color = colorFor(percent)
    const [filled, empty] = bar(percent)
    const usd = costUsd(usage?.cost)

    const api = usage?.context?.apiUsage
    const tokensIn =
      (Number(api?.input_tokens) || 0) +
      (Number(api?.cache_creation_input_tokens) || 0) +
      (Number(api?.cache_read_input_tokens) || 0)
    const tokensOut = Number(api?.output_tokens) || 0

    const parts = [
      Text({ color, children: [filled] }),
      Text({ dimColor: true, children: [empty] }),
      Text({ color, children: [' ' + percent.toFixed(1) + '%'] }),
      Text({
        children: [' (' + fmtTokens(usage?.context?.tokens) + '/' + fmtTokens(usage?.context?.window) + ')'],
      }),
    ]
    parts.push(
      Text({
        children: [' · in ' + fmtTokens(tokensIn) + ' · out ' + fmtTokens(tokensOut)],
      }),
    )
    if (usd !== null) parts.push(Text({ children: [' · ~$' + usd.toFixed(3)] }))

    // Modo a la izquierda, uso a la derecha
    // Sin width: un Box con un nodo del motor dentro no admite esa prop, asi que
    // un espaciador con flexGrow empuja el uso a la derecha.
    const top = Box({
      flexDirection: 'row',
      children: [theirs, Box({ flexGrow: 1 }), Box({ flexDirection: 'row', children: parts })],
    })
    const limits = limitsRow(Box, Text, usage?.rateLimits)
    if (!limits) return top
    return Box({ flexDirection: 'column', children: [top, limits] })
  })
}
