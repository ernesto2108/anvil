import { useState } from 'react'
import { AppShell } from '@/components/layout/app-shell'
import { RunsView } from '@/views/runs-view'
import { FlowView } from '@/views/flow-view'
import { AgentView } from '@/views/agent-view'
import { MetricsView } from '@/views/metrics-view'

type View =
  | { name: 'runs' }
  | { name: 'flow'; runId: string }
  | { name: 'agent'; runId: string; agentId: string }
  | { name: 'metrics' }

function App() {
  const [view, setView] = useState<View>({ name: 'runs' })

  const activeNav =
    view.name === 'flow' || view.name === 'agent' ? 'runs' : (view.name as 'runs' | 'metrics')

  const title =
    view.name === 'runs'
      ? 'Runs'
      : view.name === 'flow'
        ? 'Flow'
        : view.name === 'metrics'
          ? 'Metrics'
          : 'Detalle de agente'

  return (
    <AppShell
      title={title}
      activeView={activeNav}
      onNavigate={(v) => {
        if (v === 'metrics') setView({ name: 'metrics' })
        else setView({ name: 'runs' })
      }}
    >
      {view.name === 'runs' && (
        <RunsView onRowClick={(id) => setView({ name: 'flow', runId: id })} />
      )}
      {view.name === 'flow' && (
        <FlowView
          runId={view.runId}
          onBack={() => setView({ name: 'runs' })}
          onAgentSelect={(agentId) => setView({ name: 'agent', runId: view.runId, agentId })}
        />
      )}
      {view.name === 'agent' && (
        <AgentView
          runId={view.runId}
          agentId={view.agentId}
          onBack={() => setView({ name: 'flow', runId: view.runId })}
        />
      )}
      {view.name === 'metrics' && <MetricsView />}
    </AppShell>
  )
}

export default App
