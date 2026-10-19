// Pruebas de la logica pura. Se corren con: node --test mods/agent-progress/tests/progress.test.mjs
import assert from 'node:assert/strict'
import { test } from 'node:test'
import { bar, createTracker, describe, fmtElapsed, formatLine, percentOf } from '../hooks/progress.mjs'

const todos = (...statuses) => ({ todos: statuses.map((status, i) => ({ content: 't' + i, status, activeForm: 'a' + i })) })

test('cuenta llamadas y fija el inicio en la primera', () => {
  const t = createTracker()
  t.record('a1', 'Read', {}, 1000)
  t.record('a1', 'Grep', {}, 5000)
  assert.deepEqual(t.get('a1'), { startedAt: 1000, calls: 2, todos: null, lastTool: 'Grep' })
})

test('TodoWrite fija completadas y total', () => {
  const t = createTracker()
  t.record('a1', 'TodoWrite', todos('completed', 'completed', 'in_progress', 'pending'), 0)
  assert.deepEqual(t.get('a1').todos, { done: 2, total: 4 })
  assert.equal(percentOf(t.get('a1').todos), 50)
})

test('un TodoWrite nuevo reemplaza al anterior y una lista vacia se ignora', () => {
  const t = createTracker()
  t.record('a1', 'TodoWrite', todos('pending', 'pending'), 0)
  t.record('a1', 'TodoWrite', todos('completed', 'pending'), 0)
  assert.deepEqual(t.get('a1').todos, { done: 1, total: 2 })
  t.record('a1', 'TodoWrite', { todos: [] }, 0)
  assert.deepEqual(t.get('a1').todos, { done: 1, total: 2 })
})

test('los agentes se siguen por separado y se pueden olvidar', () => {
  const t = createTracker()
  t.record('a1', 'Read', {}, 0)
  t.record('a2', 'Read', {}, 0)
  t.forget('a1')
  assert.equal(t.get('a1'), undefined)
  assert.equal(t.get('a2').calls, 1)
})

test('sin plan no hay porcentaje', () => {
  assert.equal(percentOf(null), null)
  const d = describe({ label: 'x', state: { startedAt: 0, calls: 1, todos: null }, now: 42000 })
  assert.equal(d.percent, null)
  assert.equal(d.bar, '')
  assert.equal(d.plan, 'sin plan')
  assert.equal(d.rest, '1 llamada · 42s')
})

test('con plan muestra barra, porcentaje y fraccion', () => {
  const d = describe({ label: 'x', state: { startedAt: 0, calls: 14, todos: { done: 3, total: 7 } }, now: 402000 })
  assert.equal(d.percent, 43)
  assert.equal(d.bar, '████░░░░░░')
  assert.equal(d.plan, '43% (3/7)')
  assert.equal(d.rest, '14 llamadas · 6m 42s')
})

test('bar y fmtElapsed', () => {
  assert.equal(bar(0), '░░░░░░░░░░')
  assert.equal(bar(100), '██████████')
  assert.equal(bar(150), '██████████')
  assert.equal(fmtElapsed(0), '0s')
  assert.equal(fmtElapsed(65000), '1m 05s')
})

test('formatLine: con y sin plan', () => {
  const sin = describe({ label: 'Prueba', state: { startedAt: 0, calls: 12, todos: null }, now: 28000 })
  assert.equal(formatLine(sin, 'WebSearch'), 'Prueba  sin plan · 12 llamadas · 28s · WebSearch')
  const con = describe({ label: 'Prueba', state: { startedAt: 0, calls: 2, todos: { done: 1, total: 4 } }, now: 5000 })
  assert.equal(formatLine(con), 'Prueba  ███░░░░░░░ 25% (1/4) · 2 llamadas · 5s')
})
