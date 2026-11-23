// commit-es: rechaza `git commit` y `gh pr create/edit` cuyo mensaje no este en
// espanol, o cuyo asunto de commit no siga Conventional Commits.
//
// SIN VERIFICAR EN VIVO (requiere Claude Code v2.1.287+). Pendientes:
//   - que $.session.cwd() devuelva el cwd como string
//   - forma del retorno de $.fs.read (se acepta string u objeto con content/text)
// La logica vive en check.mjs y se prueba con `node --test` (ver tests/).
//
// Falla abierto a proposito: si este hook falla, el commit pasa. Es un guardia
// de estilo, no de seguridad; no debe poder bloquear tu trabajo.

import { analyze, evaluate, resolvePath } from './check.mjs'

function asText(read) {
  if (typeof read === 'string') return read
  return read?.content ?? read?.text ?? ''
}

export function register(on) {
  on('tool.call', { tool: 'Bash' }, async ($, e, next) => {
    const items = analyze(e.command)
    // Cualquier otro comando pasa sin revisar
    if (items.length === 0) return next(e)

    let cwd = ''
    try {
      cwd = await $.session.cwd()
    } catch {
      // Sin cwd, los archivos relativos de -F no se pueden leer y se omiten
    }

    const problems = []
    for (const item of items) {
      const texts = [...item.texts]
      for (const file of item.files) {
        const abs = resolvePath(cwd, file.path)
        if (!abs) continue
        try {
          texts.push({ label: file.label, text: asText(await $.fs.read(abs)) })
        } catch {
          // No se pudo leer el archivo: ante la duda, no bloquear
        }
      }
      problems.push(...evaluate(item.kind, texts))
    }

    if (problems.length === 0) return next(e)

    return {
      deny:
        'commit-es rechazó este comando. ' +
        problems.join(' ') +
        ' Reescríbelo en español (los tipos de Conventional Commits, identificadores, paths, flags y nombres propios se dejan tal cual) y vuelve a intentarlo.',
    }
  })
}
