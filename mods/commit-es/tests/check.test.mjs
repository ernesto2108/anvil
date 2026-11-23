// Pruebas de la logica pura. Se corren con: node --test mods/commit-es/tests/check.test.mjs
import assert from 'node:assert/strict'
import { test } from 'node:test'
import { analyze, evaluate, extractHeredocs, isEnglish, languageScore, resolvePath, stripNonProse } from '../hooks/check.mjs'

// Atajo: analiza un comando y evalua el primer hallazgo
const problemsOf = (command) => {
  const [item] = analyze(command)
  return item ? evaluate(item.kind, item.texts) : null
}

test('extrae mensajes de -m, varios -m y banderas agrupadas', () => {
  assert.deepEqual(analyze('git commit -m "feat: agrega login"')[0].texts, [{ label: 'mensaje', text: 'feat: agrega login' }])
  assert.deepEqual(
    analyze('git commit -m "feat: asunto" -m "cuerpo"')[0].texts.map((t) => t.text),
    ['feat: asunto', 'cuerpo'],
  )
  assert.deepEqual(analyze('git commit -am "fix: corrige bug"')[0].texts.map((t) => t.text), ['fix: corrige bug'])
  assert.deepEqual(analyze('git commit --message="fix: algo"')[0].texts.map((t) => t.text), ['fix: algo'])
  assert.deepEqual(analyze('git commit -m"fix: pegado"')[0].texts.map((t) => t.text), ['fix: pegado'])
})

test('extrae el mensaje de -F - con heredoc', () => {
  const cmd = "git commit -q -F - <<'EOF'\nfeat(mods): agrega mod\n\nCuerpo del commit.\n\nCo-Authored-By: Claude <noreply@anthropic.com>\nEOF"
  const [item] = analyze(cmd)
  assert.equal(item.kind, 'commit')
  assert.match(item.texts[0].text, /^feat\(mods\): agrega mod/)
})

test('extrae el mensaje de -m "$(cat <<EOF ...)"', () => {
  const cmd = 'git commit -m "$(cat <<\'EOF\'\nfeat: agrega algo\n\nDetalle en español.\nEOF\n)"'
  const [item] = analyze(cmd)
  assert.equal(item.texts[0].text.trim().split('\n')[0], 'feat: agrega algo')
})

test('un -F con archivo se devuelve como archivo por leer', () => {
  assert.deepEqual(analyze('git commit -F msg.txt')[0].files, [{ label: 'mensaje', path: 'msg.txt' }])
})

test('commits en cadena: git add && git commit', () => {
  const items = analyze('git add . && git commit -m "feat: agrega x"')
  assert.equal(items.length, 1)
  assert.equal(items[0].kind, 'commit')
})

test('no analiza comandos que no son commit ni PR', () => {
  assert.deepEqual(analyze('git status'), [])
  assert.deepEqual(analyze('git push origin main'), [])
  assert.deepEqual(analyze('ls -la'), [])
  assert.deepEqual(analyze('gh pr view 12'), [])
})

test('commit en español con Conventional Commits pasa', () => {
  assert.deepEqual(problemsOf('git commit -m "feat(mods): agrega mod blast-radius para comandos destructivos"'), [])
  assert.deepEqual(problemsOf('git commit -m "fix: corrige el cálculo del costo de la sesión"'), [])
  assert.deepEqual(problemsOf('git commit -m "chore(skills): poda contenido generico de go-conventions"'), [])
})

test('commit en inglés se rechaza', () => {
  const p = problemsOf('git commit -m "feat: add login endpoint"')
  assert.equal(p.length, 1)
  assert.match(p[0], /inglés/)
  assert.match(p[0], /add/)
  assert.equal(problemsOf('git commit -m "chore: update dependencies"').length, 1)
  assert.equal(problemsOf('git commit -m "fix: handle null token when the user logs out"').length, 1)
})

test('asunto sin Conventional Commits se rechaza', () => {
  const p = problemsOf('git commit -m "agrega login al sistema"')
  assert.equal(p.length, 1)
  assert.match(p[0], /Conventional Commits/)
  assert.equal(problemsOf('git commit -m "feat agrega login"').length, 1)
  assert.equal(problemsOf('git commit -m "feat:agrega login"').length, 1)
})

test('inglés y sin Conventional Commits reporta ambos problemas', () => {
  assert.equal(problemsOf('git commit -m "add login endpoint for the users"').length, 2)
})

test('el trailer Co-Authored-By y el pie de Claude Code no cuentan como inglés', () => {
  const cmd = "git commit -F - <<'EOF'\nfeat: agrega el mod de uso\n\nCo-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>\nEOF"
  assert.deepEqual(problemsOf(cmd), [])
})

test('merge, revert y fixup se omiten', () => {
  assert.deepEqual(problemsOf('git commit -m "Merge branch main into develop"'), [])
  assert.deepEqual(problemsOf('git commit -m "fixup! feat: agrega x"'), [])
})

test('git commit sin mensaje explícito no se revisa', () => {
  assert.deepEqual(problemsOf('git commit --amend --no-edit'), [])
  assert.deepEqual(problemsOf('git commit'), [])
})

test('PR: título y cuerpo en español pasan; en inglés se rechazan', () => {
  assert.deepEqual(problemsOf('gh pr create --title "Agrega el mod anvil-usage" --body "Resumen de los cambios."'), [])
  const p = problemsOf('gh pr create --title "Add the usage mod" --body "Resumen de los cambios."')
  assert.equal(p.length, 1)
  assert.match(p[0], /título/)
})

test('PR con cuerpo por heredoc y pie de Claude Code', () => {
  const cmd =
    'gh pr create --title "Agrega mods" --body "$(cat <<\'EOF\'\n## Resumen\n- Agrega el mod de uso\n\n🤖 Generated with [Claude Code](https://claude.com/claude-code)\nEOF\n)"'
  assert.deepEqual(problemsOf(cmd), [])
  const bad = 'gh pr create --title "Agrega mods" --body "$(cat <<\'EOF\'\n## Summary\n- Add the usage mod for the session\nEOF\n)"'
  assert.equal(problemsOf(bad).length, 1)
})

test('ignora identificadores, paths, código y URLs al medir el idioma', () => {
  assert.equal(isEnglish('actualiza `add_user` en src/add/user.ts y https://the.example.com/add'), false)
  assert.equal(stripNonProse('feat(auth): agrega login').trim(), 'agrega login')
})

test('un texto sin palabras reconocibles no se marca', () => {
  assert.equal(isEnglish('feat: login'), false)
  assert.deepEqual(languageScore('xyz qwerty'), { en: 0, es: 0, enWords: [] })
})

test('las tildes y la ñ cuentan como español', () => {
  assert.equal(isEnglish('the sesión añadida'), false)
})

test('extractHeredocs deja marcadores y guarda los cuerpos', () => {
  const { sanitized, bodies } = extractHeredocs("cat <<'EOF'\nhola\nEOF")
  assert.deepEqual(bodies, ['hola'])
  assert.match(sanitized, /__HEREDOC_0__/)
})

test('resolvePath', () => {
  assert.equal(resolvePath('/a/b', 'm.txt'), '/a/b/m.txt')
  assert.equal(resolvePath('/a/b', '/x/m.txt'), '/x/m.txt')
  assert.equal(resolvePath('', 'm.txt'), null)
})
