// Logica pura de agent-progress: sigue lo que hace cada subagente y arma el
// texto de su linea. Se prueba con `node --test` sin el API de mods.
//
// El porcentaje NO es una estimacion: sale de la lista de tareas (TodoWrite)
// que el subagente mantiene, completadas / total. Un subagente que no usa
// TodoWrite no tiene plan que medir, asi que solo se muestran llamadas y tiempo.

export const BAR_WIDTH = 10

export function createTracker() {
  const agents = new Map()

  return {
    // Una llamada de herramienta hecha por el subagente `agentId`
    record(agentId, tool, args, now) {
      let a = agents.get(agentId)
      if (!a) {
        a = { startedAt: now, calls: 0, todos: null, lastTool: '' }
        agents.set(agentId, a)
      }
      a.calls += 1
      a.lastTool = String(tool ?? '')
      if (tool === 'TodoWrite' && Array.isArray(args?.todos) && args.todos.length > 0) {
        const done = args.todos.filter((t) => t && t.status === 'completed').length
        a.todos = { done, total: args.todos.length }
      }
    },
    get(agentId) {
      return agents.get(agentId)
    },
    forget(agentId) {
      agents.delete(agentId)
    },
    clear() {
      agents.clear()
    },
  }
}

// 0.43 -> "████░░░░░░"
export function bar(percent, width = BAR_WIDTH) {
  const p = Math.min(100, Math.max(0, Number(percent) || 0))
  const full = Math.round((p * width) / 100)
  return '█'.repeat(full) + '░'.repeat(width - full)
}

// 402000 -> "6m 42s", 42000 -> "42s"
export function fmtElapsed(ms) {
  const s = Math.max(0, Math.floor((Number(ms) || 0) / 1000))
  const m = Math.floor(s / 60)
  return m > 0 ? `${m}m ${String(s % 60).padStart(2, '0')}s` : `${s}s`
}

export function percentOf(todos) {
  if (!todos || todos.total <= 0) return null
  return Math.round((todos.done / todos.total) * 100)
}

// Verde < 60 → en curso, amarillo < 100, verde al terminar no aplica (la fila
// desaparece al completarse). Solo orienta; el numero manda.
export function colorFor(percent) {
  return percent === null ? undefined : percent >= 80 ? 'green' : percent >= 40 ? 'yellow' : 'cyan'
}

// Devuelve las partes de la linea para que register.js las pinte
export function describe({ label, state, now }) {
  const percent = percentOf(state.todos)
  const calls = `${state.calls} ${state.calls === 1 ? 'llamada' : 'llamadas'}`
  const elapsed = fmtElapsed(now - state.startedAt)
  return {
    label,
    percent,
    bar: percent === null ? '' : bar(percent),
    plan: percent === null ? 'sin plan' : `${percent}% (${state.todos.done}/${state.todos.total})`,
    rest: `${calls} · ${elapsed}`,
  }
}

// La linea de estado es texto plano: una sola cadena por subagente
// "Prueba 2  sin plan · 12 llamadas · 28s · WebSearch"
export function formatLine(d, lastTool = '') {
  const plan = d.bar ? `${d.bar} ${d.plan}` : d.plan
  return [`${d.label}  ${plan}`, d.rest, lastTool].filter(Boolean).join(' · ')
}
