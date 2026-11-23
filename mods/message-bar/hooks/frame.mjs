// Arma los arboles de message-bar. Son funciones puras que reciben los
// elementos ya resueltos (Box, Text), para probarlas con `node --test` sin el
// API de mods.

// Clave del tema: se adapta a tema claro u oscuro. Una terminal no tiene
// opacidad real, asi que "un poco opaco" se logra con el fondo del tema.
export const MESSAGE_BG = 'userMessageBackground'
// Mas lineas que esto y se deja el dibujo original (conserva expandir/contraer)
export const MAX_LINES = 12

// 1737000000000 -> "11:26" (hora local, 24 h)
export function formatTime(ms) {
  const d = new Date(ms)
  return String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
}

// Solo se redibuja el mensaje que escribes tu y es corto. Los mensajes de
// tareas u otros agentes (props.task / props.from) y los largos conservan el
// dibujo original de Claude Code.
export function shouldRedraw(props) {
  if (!props || typeof props.text !== 'string') return false
  if (props.task || props.from) return false
  return props.text.split('\n').length <= MAX_LINES
}

// Envuelve el dibujo original con aire arriba/abajo.
export function wrapEngine(Box, message) {
  return Box({
    flexDirection: 'column',
    marginTop: 1,
    marginBottom: 1,
    children: [message],
  })
}

// Dibuja el mensaje propio: linea tenue con nombre y hora, y el texto con
// un fondo que cubre todo el ancho de la terminal (el cuerpo no usa alignSelf,
// asi que se estira).
export function buildMessage(Box, Text, { name, time, text }, colors = {}) {
  const bg = colors.bg ?? MESSAGE_BG
  const header = [name, time].filter(Boolean).join(' · ')
  const lines = text.split('\n').map((line, i) => (i === 0 ? '❯ ' : '  ') + (line || ' '))

  return Box({
    flexDirection: 'column',
    marginTop: 1,
    marginBottom: 1,
    marginRight: 1,
    children: [
      Text({ key: 'header', dimColor: true, children: [header] }),
      Box({
        key: 'body',
        flexDirection: 'column',
        backgroundColor: bg,
        paddingLeft: 2,
        paddingRight: 2,
        paddingTop: 1,
        paddingBottom: 1,
        children: lines.map((line, i) => Text({ key: 'line-' + i, children: [line] })),
      }),
    ],
  })
}
