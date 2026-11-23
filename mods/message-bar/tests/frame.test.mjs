// Pruebas de la logica pura. Se corren con: node --test mods/message-bar/tests/frame.test.mjs
import assert from 'node:assert/strict'
import { test } from 'node:test'
import { MAX_LINES, MESSAGE_BG, buildMessage, formatTime, shouldRedraw, wrapEngine } from '../hooks/frame.mjs'

// Elementos de mentira: devuelven sus props, para inspeccionar el arbol
const Box = (props) => props
const Text = (props) => ({ _text: true, ...props })

const textsOf = (node) => (node.children ?? []).flatMap((c) => (c._text ? [c.children[0]] : textsOf(c)))

test('wrapEngine envuelve el mensaje con aire arriba y abajo', () => {
  const message = { type: 'engine' }
  const tree = wrapEngine(Box, message)
  assert.equal(tree.marginTop, 1)
  assert.equal(tree.marginBottom, 1)
  assert.equal(tree.children[0], message)
})

test('buildMessage: encabezado con nombre y hora, y texto con fondo a todo el ancho', () => {
  const tree = buildMessage(Box, Text, { name: 'ernesto', time: '11:26', text: 'hola' })
  const [header, body] = tree.children
  assert.equal(header.dimColor, true)
  assert.deepEqual(header.children, ['ernesto · 11:26'])
  // El cuerpo se estira: el fondo ocupa todo el ancho
  assert.equal(body.backgroundColor, MESSAGE_BG)
  assert.equal(body.alignSelf, undefined)
  // Margen interno a los costados y arriba/abajo
  assert.deepEqual([body.paddingLeft, body.paddingRight, body.paddingTop, body.paddingBottom], [2, 2, 1, 1])
  assert.equal(tree.marginRight, 1)
  // Sin linea en blanco entre el encabezado y el fondo
  assert.equal(body.marginTop, undefined)
  assert.deepEqual(textsOf(body), ['❯ hola'])
})

test('buildMessage: varias lineas, sangria y linea vacia', () => {
  const tree = buildMessage(Box, Text, { name: 'x', time: '09:05', text: 'uno\n\ndos' })
  assert.deepEqual(textsOf(tree.children[1]), ['❯ uno', '   ', '  dos'])
})

test('buildMessage: sin nombre solo muestra la hora', () => {
  const tree = buildMessage(Box, Text, { name: '', time: '11:26', text: 'hola' })
  assert.deepEqual(tree.children[0].children, ['11:26'])
})

test('buildMessage: el color de fondo se puede cambiar', () => {
  const tree = buildMessage(Box, Text, { name: 'x', time: '1', text: 'a' }, { bg: 'gray' })
  assert.equal(tree.children[1].backgroundColor, 'gray')
})

test('buildMessage: las claves de los hijos son distintas', () => {
  const tree = buildMessage(Box, Text, { name: 'x', time: '1', text: 'a\nb' })
  const keys = [...tree.children.map((c) => c.key), ...tree.children[1].children.map((c) => c.key)]
  assert.equal(new Set(keys).size, keys.length)
})

test('shouldRedraw: solo mensajes propios y cortos', () => {
  assert.equal(shouldRedraw({ text: 'hola' }), true)
  assert.equal(shouldRedraw({ text: 'hola', origin: 'user' }), true)
  assert.equal(shouldRedraw({ text: 'a\n'.repeat(MAX_LINES - 1) + 'z' }), true)
  assert.equal(shouldRedraw({ text: 'a\n'.repeat(MAX_LINES) + 'z' }), false)
  assert.equal(shouldRedraw({ text: 'hola', task: { id: 1 } }), false)
  assert.equal(shouldRedraw({ text: 'hola', from: 'otro-agente' }), false)
  assert.equal(shouldRedraw({}), false)
  assert.equal(shouldRedraw(undefined), false)
})

test('formatTime usa hora local de 24 h con ceros', () => {
  const d = new Date(2026, 9, 6, 9, 5)
  assert.equal(formatTime(d.getTime()), '09:05')
  assert.equal(formatTime(new Date(2026, 9, 6, 23, 59).getTime()), '23:59')
})
