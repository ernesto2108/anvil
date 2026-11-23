// blast-radius: retiene los comandos Bash destructivos, muestra que cambiarian
// y pide confirmacion antes de dejarlos correr.
//
// SIN VERIFICAR EN VIVO (requiere Claude Code v2.1.287+). Pendientes:
//   - que $.process.run acepte la forma ['cmd', ...args] tal como en la doc
//   - que $.session.cwd() devuelva el cwd como string
//   - formato de la pregunta de $.ui.ask con varias lineas
// La deteccion vive en detect.mjs y se prueba con `node --test` (ver tests/).
//
// Es una red de seguridad, no un sistema de permisos: el chequeo normal de
// permisos de Claude Code corre despues de este hook.

import { capLines, classify, formatKb, hasGlob, plural, resolvePath } from './detect.mjs'

const MAX_TARGETS = 5
const MAX_LINES = 8

async function runGit($, dir, args) {
  return $.process.run(['git', '-C', dir, ...args])
}

async function describeRm($, finding, cwd, home) {
  const lines = []
  for (const target of finding.targets.slice(0, MAX_TARGETS)) {
    if (hasGlob(target)) {
      lines.push('  ' + target + ': patrón, no se expande aquí')
      continue
    }
    const abs = resolvePath(cwd, target, home)
    if (!abs) {
      lines.push('  ' + target + ': no se pudo resolver la ruta')
      continue
    }
    const du = await $.process.run(['du', '-sk', '--', abs])
    if (du.exitCode !== 0) {
      lines.push('  ' + target + ': no existe')
      continue
    }
    const kb = parseInt(du.stdout, 10) || 0
    const found = await $.process.run(['find', abs, '-type', 'f'])
    const count = found.exitCode === 0 ? found.stdout.split('\n').filter(Boolean).length : '?'
    const files = typeof count === 'number' ? plural(count, 'archivo', 'archivos') : '? archivos'
    lines.push('  ' + target + ': ' + files + ', ' + formatKb(kb))
  }
  if (finding.targets.length > MAX_TARGETS) lines.push('  … y ' + (finding.targets.length - MAX_TARGETS) + ' objetivos más')
  if (finding.targets.length === 0) lines.push('  (sin objetivos detectados)')
  return ['Eliminaría:', ...lines]
}

async function describeGit($, finding, cwd, home) {
  const dir = finding.dir ? resolvePath(cwd, finding.dir, home) : cwd
  if (!dir) return ['  (no se pudo resolver el directorio de git)']

  if (finding.kind === 'reset-hard') {
    const out = await runGit($, dir, ['status', '--short'])
    // Los archivos sin seguimiento (??) no los toca reset --hard
    const changed = out.stdout.split('\n').filter((l) => l && !l.startsWith('??'))
    if (out.exitCode !== 0) return ['Descartaría cambios (git reset --hard):', '  no se pudo leer git status']
    if (changed.length === 0) return ['Descartaría cambios (git reset --hard):', '  no hay cambios sin commitear']
    return ['Descartaría ' + plural(changed.length, 'cambio', 'cambios') + ' sin commitear:',...capLines(changed, MAX_LINES).map((l) => '  ' + l)]
  }

  if (finding.kind === 'discard') {
    const out = await runGit($, dir, ['diff', '--name-only'])
    const files = out.stdout.split('\n').filter(Boolean)
    if (out.exitCode !== 0) return ['Descartaría cambios sin stagear:', '  no se pudo leer git diff']
    if (files.length === 0) return ['Descartaría cambios sin stagear:', '  no hay cambios sin stagear']
    return ['Descartaría cambios sin stagear en ' + plural(files.length, 'archivo', 'archivos') + ':',...capLines(files, MAX_LINES).map((l) => '  ' + l)]
  }

  if (finding.kind === 'clean') {
    const out = await runGit($, dir, ['clean', '-nd'])
    const paths = out.stdout.split('\n').filter(Boolean)
    if (out.exitCode !== 0) return ['Borraría archivos sin seguimiento:', '  no se pudo simular git clean']
    if (paths.length === 0) return ['Borraría archivos sin seguimiento:', '  no hay nada que borrar']
    return ['Borraría ' + plural(paths.length, 'ruta', 'rutas') + ' sin seguimiento:',...capLines(paths, MAX_LINES).map((l) => '  ' + l)]
  }

  // force-push: commits remotos que no existen localmente y se perderian
  const out = await runGit($, dir, ['log', '--oneline', 'HEAD..@{u}'])
  const commits = out.stdout.split('\n').filter(Boolean)
  if (out.exitCode !== 0) return ['Force push:', '  sin upstream configurado, no pude calcular qué commits se perderían']
  if (commits.length === 0) return ['Force push:', '  no se perdería ningún commit remoto']
  return ['Se perdería' + (commits.length === 1 ? '' : 'n') + ' ' + plural(commits.length, 'commit remoto', 'commits remotos') + ':',...capLines(commits, MAX_LINES).map((l) => '  ' + l)]
}

export function register(on) {
  on('tool.call', { tool: 'Bash' }, async ($, e, next) => {
    const findings = classify(e.command)
    // Cualquier otro comando pasa sin preguntar
    if (findings.length === 0) return next(e)

    let cwd = ''
    let home = ''
    try {
      cwd = await $.session.cwd()
      home = (await $.env.get('HOME')) ?? ''
    } catch {
      // Sin cwd las rutas relativas se reportan como "no se pudo resolver"
    }

    const sections = []
    for (const finding of findings) {
      try {
        const lines = finding.kind === 'rm' ? await describeRm($, finding, cwd, home) : await describeGit($, finding, cwd, home)
        sections.push(...lines)
      } catch (err) {
        sections.push('  no se pudo calcular el alcance: ' + (err?.message ?? err))
      }
    }

    const question = ['Comando con efecto destructivo:', '  ' + e.command, '', ...sections, '', '¿Ejecutarlo?'].join('\n')

    // Parte de la respuesta segura: si nadie responde, se cancela
    let answer = 'Cancelar'
    try {
      // La llamada queda retenida hasta que el usuario elige
      answer = await $.ui.ask(question, ['Proceder', 'Cancelar'])
    } catch {
      // El usuario descartó la pregunta, o no hay nadie (claude -p)
    }
    if (answer !== 'Proceder') {
      return { deny: 'El usuario canceló este comando. Pregunta antes de intentar otra alternativa.' }
    }
    // El chequeo normal de permisos sigue corriendo después
    return next(e)
  }).catch(async ($, e, next) => {
    // Falla cerrado: si este hook falla, el comando no corre
    return { deny: 'blast-radius falló (' + next.error.kind + '), el comando no se ejecutó.' }
  })
}
