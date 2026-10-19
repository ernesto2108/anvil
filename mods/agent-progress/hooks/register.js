// agent-progress: una linea por subagente en ejecucion con su avance, debajo de
// la fila de modo ("auto mode on ... <- 2 agents"), pegada a la lista de agentes.
// El porcentaje es real solo cuando el subagente mantiene una lista de tareas
// (TodoWrite): completadas / total. Si no la usa, muestra "sin plan" con
// llamadas, tiempo y la ultima herramienta, que es lo unico que se mide de verdad.
//
// La fila de cada agente en el panel ("general-purpose ... 33s . tokens") no es
// un componente que los mods puedan dibujar en esta version, por eso se dibuja
// en SessionMode. Verificado en vivo: tool.call trae agentId del subagente.
// SIN VERIFICAR: que $.agent.list() devuelva status 'running' y que el Box
// column que envuelve el dibujo de otros mods (anvil-usage) sea valido.
// Los contadores viven en variables del modulo: se reinician al recargar el mod.

import { createTracker, colorFor, describe } from './progress.mjs'

const REFRESH_MS = 1000
const tracker = createTracker()

export function register(on) {
  on('session.start', async ($, e, next) => {
    // Borra la linea de estado que dejo una version anterior de este mod
    $.ui.status(undefined)
    // Claude Code no redibuja solo: refresca para que corran tiempo y llamadas
    $.clock.every(REFRESH_MS, () => $.ui.invalidate('ui.render'))
    return next(e)
  })

  on('session.end', async ($, e, next) => {
    tracker.clear()
    return next(e)
  })

  // Solo cuentan las llamadas dentro de un subagente (agentId definido)
  on('tool.call', async ($, e, next) => {
    if (e.agentId !== undefined) {
      tracker.record(e.agentId, e.tool, e, await $.clock.now())
    }
    return next(e)
  })

  on('ui.render', { component: 'SessionMode' }, async ($, e, next) => {
    // El dibujo de Claude Code (y el de otros mods, como anvil-usage)
    const theirs = await next(e)

    const list = await $.agent.list()
    for (const a of list) if (a.status !== 'running') tracker.forget(a.id)

    const now = await $.clock.now()
    const rows = list
      .filter((a) => a.status === 'running' && tracker.get(a.id))
      .map((a) => {
        const state = tracker.get(a.id)
        return { d: describe({ label: a.description || a.type, state, now }), last: state.lastTool }
      })
    if (rows.length === 0) return theirs

    const { Box, Text } = $.ui.resolve(e)
    return Box({
      flexDirection: 'column',
      children: [
        theirs,
        ...rows.map(({ d, last }, i) =>
          Box({
            key: 'agent-' + i,
            flexDirection: 'row',
            children: [
              Text({ key: 'label', children: [d.label + '  '] }),
              Text({ key: 'bar', color: colorFor(d.percent), children: [d.bar ? d.bar + ' ' : ''] }),
              Text({ key: 'plan', bold: d.percent !== null, dimColor: d.percent === null, children: [d.plan] }),
              Text({ key: 'rest', dimColor: true, children: [' · ' + d.rest + (last ? ' · ' + last : '')] }),
            ],
          })
        ),
      ],
    })
  })
}
