// message-bar: ayuda a ubicar tus mensajes en el transcript al hacer scroll.
// Dibuja una fila en blanco arriba y abajo,
// una linea tenue con tu nombre y la hora, y un fondo que cubre todo el ancho
// de la terminal.
//
// Los mensajes largos (mas de 12 lineas) y los que no son tuyos (tareas, otros
// agentes) conservan el dibujo original de Claude Code, envuelto con aire
// arriba y abajo, para no perder cosas como expandir/contraer.
//
// SIN VERIFICAR EN VIVO (requiere Claude Code v2.1.287+). Pendientes:
//   - que 'userMessageBackground' sea una clave de tema valida (ver los tipos
//     generados en .claude-plugin/types); si no, el arbol no valida y Claude
//     Code dibuja su fila normal
//   - valores de e.props.origin y si task/from identifican bien lo que no es tuyo
//   - que $.clock.now() devuelva milisegundos
//   - que flexGrow y children: [] sean validos
// La hora es la del primer dibujo del mensaje (~ el momento del envio); al
// reanudar una sesion los mensajes viejos muestran la hora de la reanudacion.
// El nombre sale de la variable de entorno MESSAGE_BAR_NAME o, si no existe, de USER.

import { buildMessage, formatTime, shouldRedraw, wrapEngine } from './frame.mjs'

// requestId del mensaje -> momento (ms) en que se dibujo por primera vez
const firstSeen = new Map()
let cachedName

export function register(on) {
  on('ui.render', { component: 'UserMessage' }, async ($, e, next) => {
    const { Box, Text } = $.ui.resolve(e)
    // El dibujo propio de Claude Code para el mensaje
    const theirs = await next(e)
    if (!shouldRedraw(e.props)) return wrapEngine(Box, theirs)

    let at = firstSeen.get(e.requestId)
    if (at === undefined) {
      at = await $.clock.now()
      firstSeen.set(e.requestId, at)
    }

    if (cachedName === undefined) {
      cachedName = (await $.env.get('MESSAGE_BAR_NAME')) || (await $.env.get('USER')) || ''
    }

    return buildMessage(Box, Text, { name: cachedName, time: formatTime(at), text: e.props.text })
  })
}
