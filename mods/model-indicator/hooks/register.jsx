// model-indicator: fila tenue arriba del textarea con el modelo activo.
// Con "opus plan / sonnet codea" el modelo cambia al entrar/salir de plan, asi
// que se toma el modelo de cada request (turn.step) y no el del ajuste.
//
// SIN VERIFICAR EN VIVO: que el atom redibuje la banda al cambiar de modelo.

import { atom, read, update } from 'claude-code'

const lastModel = atom({ plugin: 'model-indicator', key: 'lastModel' }, '')

// 'claude-sonnet-5-5' -> 'Sonnet 5.5', 'claude-opus-4-1-20250805' -> 'Opus 4.1'
export function prettyModel(id) {
  const m = /(opus|sonnet|haiku|fable)-(\d+)(?:-(\d{1,2}))?(?:-\d{8})?/i.exec(id)
  if (!m) return id
  const name = m[1][0].toUpperCase() + m[1].slice(1).toLowerCase()
  return m[3] ? `${name} ${m[2]}.${m[3]}` : `${name} ${m[2]}`
}

// Modelo real de cada request del loop principal (opusplan: opus en plan, sonnet al codear)
async function* trackStep($, e, next) {
  if (e.model && !e.agentId) await update($, lastModel, () => e.model)
  return yield* next(e)
}

export function register(on) {
  on('turn.step', trackStep)

  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    if (e.props.hasSurvey) return next(e)
    const model = (await read($, lastModel)) || (await $.session.model())
    if (!model) return next(e)
    const { Box, Text } = $.ui.resolve(e)
    return (
      <Box>
        <Text dimColor>modelo: {prettyModel(model)}</Text>
      </Box>
    )
  })
}
