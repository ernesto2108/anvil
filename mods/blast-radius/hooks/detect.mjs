// Logica pura de blast-radius: detecta comandos destructivos en un string de
// shell. No usa el API de mods ($), asi que se prueba con `node --test`.
//
// Limitaciones (es una red de seguridad, no un parser de shell):
//   - no entiende $(), backticks, alias, eval ni funciones
//   - no sigue `cd dir && ...` (solo `git -C dir`)
//   - un operador dentro de comillas no se separa, pero un heredoc si se parte mal

// Parte un comando compuesto en segmentos por ; && || | & y saltos de linea,
// respetando comillas simples y dobles.
export function splitCommand(command) {
  const segments = []
  let cur = ''
  let quote = null
  for (let i = 0; i < command.length; i++) {
    const c = command[i]
    if (quote) {
      if (c === quote) quote = null
      else if (c === '\\' && quote === '"' && i + 1 < command.length) {
        cur += c + command[++i]
        continue
      }
      cur += c
      continue
    }
    if (c === '"' || c === "'") {
      quote = c
      cur += c
      continue
    }
    if (c === '\\' && i + 1 < command.length) {
      cur += c + command[++i]
      continue
    }
    if (c === ';' || c === '|' || c === '&' || c === '\n') {
      if ((c === '|' || c === '&') && command[i + 1] === c) i++
      if (cur.trim()) segments.push(cur.trim())
      cur = ''
      continue
    }
    cur += c
  }
  if (cur.trim()) segments.push(cur.trim())
  return segments
}

// Tokeniza un segmento al estilo shell: espacios separan, las comillas agrupan.
export function tokenize(segment) {
  const tokens = []
  let cur = ''
  let quote = null
  let has = false
  for (let i = 0; i < segment.length; i++) {
    const c = segment[i]
    if (quote) {
      if (c === quote) quote = null
      else if (c === '\\' && quote === '"' && i + 1 < segment.length) cur += segment[++i]
      else cur += c
      continue
    }
    if (c === '"' || c === "'") {
      quote = c
      has = true
      continue
    }
    if (c === '\\' && i + 1 < segment.length) {
      cur += segment[++i]
      has = true
      continue
    }
    if (/\s/.test(c)) {
      if (cur || has) {
        tokens.push(cur)
        cur = ''
        has = false
      }
      continue
    }
    cur += c
  }
  if (cur || has) tokens.push(cur)
  return tokens
}

const WRAPPERS = new Set(['sudo', 'command', 'nohup', 'time'])

// Quita envoltorios (sudo...) y asignaciones de entorno (FOO=bar) del inicio.
function stripPrefix(tokens) {
  let i = 0
  while (i < tokens.length && (WRAPPERS.has(tokens[i]) || /^[A-Za-z_][A-Za-z0-9_]*=/.test(tokens[i]))) i++
  return tokens.slice(i)
}

function classifyRm(args) {
  const destructive = args.some((a) => a === '--recursive' || a === '--force' || /^-[a-zA-Z]*[rRf][a-zA-Z]*$/.test(a))
  if (!destructive) return null
  const targets = []
  let afterDashes = false
  for (const a of args) {
    if (!afterDashes && a === '--') {
      afterDashes = true
      continue
    }
    if (!afterDashes && a.startsWith('-')) continue
    targets.push(a)
  }
  return { kind: 'rm', targets }
}

function classifyGit(tokens) {
  let i = 1
  let dir = null
  while (i < tokens.length && tokens[i].startsWith('-')) {
    if (tokens[i] === '-C' && i + 1 < tokens.length) {
      dir = tokens[i + 1]
      i += 2
    } else if (['-c', '--git-dir', '--work-tree'].includes(tokens[i]) && i + 1 < tokens.length) {
      i += 2
    } else {
      i++
    }
  }
  const sub = tokens[i]
  const args = tokens.slice(i + 1)

  if (sub === 'reset' && args.includes('--hard')) return { kind: 'reset-hard', dir }

  if (sub === 'checkout') {
    const dashes = args.indexOf('--')
    if (dashes >= 0 && args.slice(dashes + 1).includes('.')) return { kind: 'discard', dir }
    return null
  }

  if (sub === 'restore') {
    const staged = args.includes('--staged') || args.includes('-S')
    if (!staged && args.includes('.')) return { kind: 'discard', dir }
    return null
  }

  if (sub === 'clean') {
    const dryRun = args.some((a) => a === '--dry-run' || /^-[a-zA-Z]*n[a-zA-Z]*$/.test(a))
    return dryRun ? null : { kind: 'clean', dir }
  }

  if (sub === 'push') {
    const forced = args.some(
      (a) =>
        a === '--force' ||
        a === '--force-with-lease' ||
        a.startsWith('--force-with-lease=') ||
        /^-[a-zA-Z]*f[a-zA-Z]*$/.test(a),
    )
    return forced ? { kind: 'force-push', dir } : null
  }
  return null
}

// Devuelve los hallazgos destructivos de un comando: [{ kind, ... }].
// kind: 'rm' | 'reset-hard' | 'discard' | 'clean' | 'force-push'
export function classify(command) {
  const findings = []
  for (const segment of splitCommand(String(command ?? ''))) {
    const tokens = stripPrefix(tokenize(segment))
    if (tokens.length === 0) continue
    let found = null
    if (tokens[0] === 'rm') found = classifyRm(tokens.slice(1))
    else if (tokens[0] === 'git') found = classifyGit(tokens)
    if (found) findings.push({ ...found, segment })
  }
  return findings
}

export function hasGlob(path) {
  return /[*?[]/.test(path)
}

// Resuelve una ruta contra cwd (y ~ contra home), normalizando . y ..
// Devuelve null si no se puede resolver.
export function resolvePath(cwd, path, home) {
  let p = path
  if (p === '~' || p.startsWith('~/')) {
    if (!home) return null
    p = home + p.slice(1)
  }
  if (!p.startsWith('/')) {
    if (!cwd) return null
    p = cwd + '/' + p
  }
  const out = []
  for (const part of p.split('/')) {
    if (part === '' || part === '.') continue
    if (part === '..') out.pop()
    else out.push(part)
  }
  return '/' + out.join('/')
}

// 900 -> "900 KB", 2048 -> "2.0 MB", 3145728 -> "3.0 GB"
export function formatKb(kb) {
  if (kb < 1024) return kb + ' KB'
  if (kb < 1024 * 1024) return (kb / 1024).toFixed(1) + ' MB'
  return (kb / (1024 * 1024)).toFixed(1) + ' GB'
}

// plural(1, 'cambio', 'cambios') -> "1 cambio"; plural(3, ...) -> "3 cambios"
export function plural(n, one, many) {
  return n + ' ' + (n === 1 ? one : many)
}

// Limita una lista de lineas y avisa cuantas quedaron fuera.
export function capLines(lines, max) {
  if (lines.length <= max) return lines
  return [...lines.slice(0, max), '… y ' + (lines.length - max) + ' más']
}
