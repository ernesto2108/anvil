import { useState } from 'react'
import { AppShell } from '@/components/layout/app-shell'
import { RunsView } from '@/views/runs-view'
import { FlowView } from '@/views/flow-view'
import { AgentView } from '@/views/agent-view'

type View =
  | { name: 'runs' }
  | { name: 'flow'; runId: string }
  | { name: 'agent'; runId: string; agentId: string }

function App() {
  const [view, setView] = useState<View>({ name: 'runs' })

  const title =
    view.name === 'runs'
      ? 'Runs'
      : view.name === 'flow'
        ? 'Flow'
        : 'Detalle de agente'

  return (
    <AppShell
      title={title}
      activeView="runs"
      onNavigate={() => setView({ name: 'runs' })}
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
    </AppShell>
  )
}

export default App
