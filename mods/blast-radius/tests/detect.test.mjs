// Pruebas de la logica pura. Se corren con: node --test mods/blast-radius/tests
import assert from 'node:assert/strict'
import { test } from 'node:test'
import { capLines, classify, formatKb, hasGlob, plural, resolvePath, splitCommand, tokenize } from '../hooks/detect.mjs'

const kinds = (command) => classify(command).map((f) => f.kind)

test('splitCommand separa por operadores y respeta comillas', () => {
  assert.deepEqual(splitCommand('a && b; c | d || e'), ['a', 'b', 'c', 'd', 'e'])
  assert.deepEqual(splitCommand('echo "a && b" && ls'), ['echo "a && b"', 'ls'])
  assert.deepEqual(splitCommand("echo 'a; b'"), ["echo 'a; b'"])
})

test('tokenize agrupa por comillas y escapes', () => {
  assert.deepEqual(tokenize('rm -rf "mi carpeta" otro'), ['rm', '-rf', 'mi carpeta', 'otro'])
  assert.deepEqual(tokenize("git commit -m 'a b'"), ['git', 'commit', '-m', 'a b'])
  assert.deepEqual(tokenize('rm a\\ b'), ['rm', 'a b'])
})

test('detecta rm recursivo o forzado y extrae objetivos', () => {
  assert.deepEqual(classify('rm -rf build dist')[0].targets, ['build', 'dist'])
  assert.deepEqual(kinds('rm -r carpeta'), ['rm'])
  assert.deepEqual(kinds('rm -f archivo'), ['rm'])
  assert.deepEqual(kinds('rm --recursive --force x'), ['rm'])
  assert.deepEqual(classify('rm -rf -- -raro')[0].targets, ['-raro'])
})

test('no marca rm simple ni comandos inocuos', () => {
  assert.deepEqual(kinds('rm archivo.txt'), [])
  assert.deepEqual(kinds('ls -la'), [])
  assert.deepEqual(kinds('git status'), [])
  assert.deepEqual(kinds('git push origin main'), [])
  assert.deepEqual(kinds('git reset --soft HEAD~1'), [])
  assert.deepEqual(kinds('git clean -n'), [])
  assert.deepEqual(kinds('git clean -nd'), [])
  assert.deepEqual(kinds('git restore --staged .'), [])
})

test('detecta operaciones destructivas de git', () => {
  assert.deepEqual(kinds('git reset --hard HEAD~1'), ['reset-hard'])
  assert.deepEqual(kinds('git checkout -- .'), ['discard'])
  assert.deepEqual(kinds('git restore .'), ['discard'])
  assert.deepEqual(kinds('git clean -fd'), ['clean'])
  assert.deepEqual(kinds('git push --force'), ['force-push'])
  assert.deepEqual(kinds('git push -f origin main'), ['force-push'])
  assert.deepEqual(kinds('git push --force-with-lease'), ['force-push'])
})

test('git -C guarda el directorio', () => {
  const [f] = classify('git -C ../otro reset --hard')
  assert.equal(f.kind, 'reset-hard')
  assert.equal(f.dir, '../otro')
})

test('detecta cada hallazgo de un comando compuesto', () => {
  assert.deepEqual(kinds('npm test && rm -rf dist; git push -f'), ['rm', 'force-push'])
})

test('ignora sudo y variables de entorno al inicio', () => {
  assert.deepEqual(kinds('sudo rm -rf /tmp/x'), ['rm'])
  assert.deepEqual(kinds('FOO=1 git reset --hard'), ['reset-hard'])
})

test('un operador dentro de comillas no cuenta como comando', () => {
  assert.deepEqual(kinds('echo "rm -rf x"'), [])
})

test('resolvePath normaliza y resuelve relativas y ~', () => {
  assert.equal(resolvePath('/a/b', 'c'), '/a/b/c')
  assert.equal(resolvePath('/a/b', '../c'), '/a/c')
  assert.equal(resolvePath('/a/b', './c/./d'), '/a/b/c/d')
  assert.equal(resolvePath('/a/b', '/x/y'), '/x/y')
  assert.equal(resolvePath('/a', '~/z', '/home/u'), '/home/u/z')
  assert.equal(resolvePath('/a', '~/z'), null)
  assert.equal(resolvePath('', 'rel'), null)
})

test('hasGlob', () => {
  assert.equal(hasGlob('*.log'), true)
  assert.equal(hasGlob('a?b'), true)
  assert.equal(hasGlob('[ab]'), true)
  assert.equal(hasGlob('normal'), false)
})

test('formatKb', () => {
  assert.equal(formatKb(900), '900 KB')
  assert.equal(formatKb(2048), '2.0 MB')
  assert.equal(formatKb(3 * 1024 * 1024), '3.0 GB')
})

test('plural', () => {
  assert.equal(plural(1, 'cambio', 'cambios'), '1 cambio')
  assert.equal(plural(0, 'cambio', 'cambios'), '0 cambios')
  assert.equal(plural(3, 'cambio', 'cambios'), '3 cambios')
})

test('capLines recorta y avisa', () => {
  assert.deepEqual(capLines(['a', 'b'], 3), ['a', 'b'])
  assert.deepEqual(capLines(['a', 'b', 'c', 'd'], 2), ['a', 'b', '… y 2 más'])
})
