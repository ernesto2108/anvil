// Logica pura de commit-es: extrae el mensaje de `git commit` y `gh pr
// create/edit` desde un comando Bash y comprueba idioma y Conventional Commits.
// No usa el API de mods ($), asi que se prueba con `node --test`.
//
// Es heuristica, no un detector de idioma real: solo marca un texto cuando
// tiene mas palabras en ingles conocidas que en espanol. Un texto sin palabras
// reconocibles pasa (ante la duda, no bloquear).
//
// splitCommand/tokenize se duplican de blast-radius a proposito: un mod solo
// puede importar archivos de su propio directorio.

// ---------- Shell: heredocs, segmentos, tokens ----------

const HEREDOC = /<<-?[ \t]*(['"]?)([A-Za-z_][A-Za-z0-9_]*)\1[ \t]*\n([\s\S]*?)\n[ \t]*\2(?=[ \t]*(?:\n|$|\)|"|'|;|&|\|))/g

// Saca los heredocs del comando y deja marcadores __HEREDOC_n__, para que sus
// lineas no se partan como si fueran comandos.
export function extractHeredocs(command) {
  const bodies = []
  const sanitized = String(command ?? '').replace(HEREDOC, (_m, _q, _tag, body) => {
    bodies.push(body)
    return ' __HEREDOC_' + (bodies.length - 1) + '__ '
  })
  return { sanitized, bodies }
}

function fillHeredocs(text, bodies) {
  return text.replace(/__HEREDOC_(\d+)__/g, (_m, i) => bodies[Number(i)] ?? '')
}

// "$(cat <cuerpo>)" -> "<cuerpo>"
function unwrapCat(text) {
  if (!text.startsWith('$(cat')) return text
  return text.replace(/^\$\(cat\s*/, '').replace(/\s*\)$/, '')
}

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

function stripPrefix(tokens) {
  let i = 0
  while (i < tokens.length && (WRAPPERS.has(tokens[i]) || /^[A-Za-z_][A-Za-z0-9_]*=/.test(tokens[i]))) i++
  return tokens.slice(i)
}

// ---------- Extraccion de mensajes ----------

// Valor del flag en la posicion i: acepta `-x valor`, `--flag valor`,
// `--flag=valor` y `-xvalor`. Devuelve { value, next } o null.
function flagValue(args, i, short, long) {
  const a = args[i]
  if (a === '--' + long) return args[i + 1] === undefined ? null : { value: args[i + 1], next: i + 2 }
  if (a.startsWith('--' + long + '=')) return { value: a.slice(long.length + 3), next: i + 1 }
  if (short) {
    // -m, -am, -sm: banderas cortas agrupadas que terminan en la letra
    if (new RegExp('^-[a-zA-Z]*' + short + '$').test(a)) {
      return args[i + 1] === undefined ? null : { value: args[i + 1], next: i + 2 }
    }
    // -mTexto pegado
    if (a.startsWith('-' + short) && a.length > 2 && !a.startsWith('--')) return { value: a.slice(2), next: i + 1 }
  }
  return null
}

function heredocRefs(tokens, bodies) {
  const out = []
  for (const t of tokens) {
    const m = /^__HEREDOC_(\d+)__$/.exec(t)
    if (m && bodies[Number(m[1])] !== undefined) out.push(bodies[Number(m[1])])
  }
  return out
}

function resolveValue(raw, bodies) {
  return unwrapCat(fillHeredocs(raw, bodies))
}

function commitFrom(tokens, bodies) {
  let i = 1
  while (i < tokens.length && tokens[i].startsWith('-')) {
    if (['-C', '-c', '--git-dir', '--work-tree'].includes(tokens[i]) && i + 1 < tokens.length) i += 2
    else i++
  }
  if (tokens[i] !== 'commit') return null
  const args = tokens.slice(i + 1)
  const texts = []
  const files = []
  let fromStdin = false
  for (let j = 0; j < args.length; ) {
    const msg = flagValue(args, j, 'm', 'message')
    if (msg) {
      texts.push({ label: 'mensaje', text: resolveValue(msg.value, bodies) })
      j = msg.next
      continue
    }
    const file = flagValue(args, j, 'F', 'file')
    if (file) {
      if (file.value === '-') fromStdin = true
      else files.push({ label: 'mensaje', path: file.value })
      j = file.next
      continue
    }
    j++
  }
  if (fromStdin) for (const body of heredocRefs(args, bodies)) texts.push({ label: 'mensaje', text: body })
  return { kind: 'commit', texts, files }
}

function prFrom(tokens, bodies) {
  if (tokens[1] !== 'pr' || !['create', 'edit'].includes(tokens[2])) return null
  const args = tokens.slice(3)
  const texts = []
  const files = []
  let fromStdin = false
  for (let j = 0; j < args.length; ) {
    const title = flagValue(args, j, 't', 'title')
    if (title) {
      texts.push({ label: 'título', text: resolveValue(title.value, bodies) })
      j = title.next
      continue
    }
    const body = flagValue(args, j, 'b', 'body')
    if (body) {
      texts.push({ label: 'cuerpo', text: resolveValue(body.value, bodies) })
      j = body.next
      continue
    }
    const file = flagValue(args, j, 'F', 'body-file')
    if (file) {
      if (file.value === '-') fromStdin = true
      else files.push({ label: 'cuerpo', path: file.value })
      j = file.next
      continue
    }
    j++
  }
  if (fromStdin) for (const text of heredocRefs(args, bodies)) texts.push({ label: 'cuerpo', text })
  return { kind: 'pr', texts, files }
}

// Devuelve [{ kind: 'commit' | 'pr', texts: [{label, text}], files: [{label, path}] }]
export function analyze(command) {
  const { sanitized, bodies } = extractHeredocs(command)
  const items = []
  for (const segment of splitCommand(sanitized)) {
    const tokens = stripPrefix(tokenize(segment))
    let item = null
    if (tokens[0] === 'git') item = commitFrom(tokens, bodies)
    else if (tokens[0] === 'gh') item = prFrom(tokens, bodies)
    if (item) items.push(item)
  }
  return items
}

// ---------- Idioma ----------

// Palabras en ingles que no son espanol. Se evitan las ambiguas (test, docs,
// build, refactor, release...) para no generar falsos positivos.
const EN = new Set(
  (
    'the and for with from this that when which into instead not are was were been being have has had ' +
    'add adds added adding fix fixes fixed fixing update updates updated updating remove removes removed removing ' +
    'improve improves improved implement implements implemented create creates created allow allows allowed ' +
    'use uses used using prevent prevents change changes changed move moves moved rename renames renamed ' +
    'handle handles handled support supports supported make makes made show shows showed resolve resolves ' +
    'bump drop drops dropped introduces enable enables disable disables ensure ensures avoid avoids ' +
    'should would could also now new only some more all any each other than then them they their there ' +
    'to of is it be as at by on or if so we you'
  ).split(' '),
)

const ES = new Set(
  (
    'de la el los las un una unos unas para con por que en se del al sin como más pero también ' +
    'agrega agregar agregan corrige corregir actualiza actualizar elimina eliminar mejora mejorar añade añadir ' +
    'implementa implementar crea crear permite permitir usa usar evita evitar cambia cambiar mueve mover ' +
    'renombra renombrar documenta documentar ajusta ajustar configura configurar soporta deja dejar quita quitar ' +
    'reemplaza reemplazar sincroniza sincronizar valida validar acopla poda extrae unifica separa habilita ' +
    'deshabilita restaura revierte optimiza simplifica introduce incluye incluir muestra mostrar ' +
    'cuando donde desde hasta entre sobre solo todo toda todos todas nuevo nueva nuevos nuevas ' +
    'este esta estos estas ese esa cada otro otra tras ante bajo según'
  ).split(' '),
)

const CC_TYPES = ['feat', 'fix', 'docs', 'style', 'refactor', 'perf', 'test', 'build', 'ci', 'chore', 'revert']
const CC_PREFIX = new RegExp('^(' + CC_TYPES.join('|') + ')(\\([^)\\n]*\\))?!?:[ \\t]*', 'gim')

// Deja solo prosa: quita codigo, URLs, trailers, el prefijo de Conventional
// Commits y los tokens que parecen identificadores, paths o numeros.
export function stripNonProse(text) {
  let t = String(text ?? '')
  t = t.replace(/```[\s\S]*?```/g, ' ').replace(/`[^`\n]*`/g, ' ')
  t = t.replace(/https?:\/\/\S+/g, ' ')
  t = t
    .split('\n')
    .filter((line) => !/^\s*(co-authored-by|signed-off-by|reviewed-by|acked-by)\s*:/i.test(line) && !/generated with/i.test(line))
    .join('\n')
  t = t.replace(CC_PREFIX, '')
  const kept = []
  for (const raw of t.split(/\s+/)) {
    const tok = raw.replace(/^[("'¿¡*\-#>]+/, '').replace(/[.,;:!?)"'*]+$/, '')
    if (!tok) continue
    if (/[/_@#=<>{}[\]\\*0-9]|\w\.\w|\w:\w|\w\(/.test(tok)) continue
    kept.push(tok)
  }
  return kept.join(' ')
}

export function languageScore(text) {
  let en = 0
  let es = 0
  const enWords = []
  for (const word of stripNonProse(text).toLowerCase().match(/[a-záéíóúñü]+/g) ?? []) {
    if (/[áéíóúñü]/.test(word)) es++
    if (EN.has(word)) {
      en++
      if (!enWords.includes(word)) enWords.push(word)
    } else if (ES.has(word)) {
      es++
    }
  }
  return { en, es, enWords }
}

// Ingles si hay mas palabras en ingles que en espanol
export function isEnglish(text) {
  const { en, es } = languageScore(text)
  return en > es
}

// ---------- Reglas ----------

const CC_SUBJECT = new RegExp('^(' + CC_TYPES.join('|') + ')(\\([^)\\n]+\\))?!?: \\S')
const SKIP_SUBJECT = /^(Merge |Revert "|fixup! |squash! |amend! )/

export function firstLine(text) {
  return String(text ?? '')
    .split('\n')
    .map((l) => l.trim())
    .find(Boolean) ?? ''
}

// Devuelve la lista de problemas (vacia si todo esta bien).
// kind: 'commit' | 'pr'; texts: [{ label, text }]
export function evaluate(kind, texts) {
  const problems = []
  const real = texts.filter((t) => String(t.text ?? '').trim())
  if (real.length === 0) return problems

  if (kind === 'commit') {
    const subject = firstLine(real[0].text)
    if (SKIP_SUBJECT.test(subject)) return problems
    if (!CC_SUBJECT.test(subject)) {
      problems.push('El asunto no sigue Conventional Commits: usa "tipo(scope): descripción" con un tipo de ' + CC_TYPES.join(', ') + '.')
    }
    const { en, es, enWords } = languageScore(real.map((t) => t.text).join('\n\n'))
    if (en > es) problems.push('El mensaje del commit está en inglés (' + enWords.slice(0, 5).join(', ') + '); debe ir en español.')
    return problems
  }

  for (const t of real) {
    const { en, es, enWords } = languageScore(t.text)
    if (en > es) problems.push('El ' + t.label + ' del PR está en inglés (' + enWords.slice(0, 5).join(', ') + '); debe ir en español.')
  }
  return problems
}

// ---------- Rutas ----------

export function resolvePath(cwd, path) {
  let p = path
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
