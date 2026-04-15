import { useCallback, useRef, useState } from 'react'
import { AppShell } from '@/components/layout/app-shell'
import { ErrorBoundary } from '@/components/error-boundary'
import { RunsView } from '@/views/runs-view'
import { FlowView } from '@/views/flow-view'
import { AgentView } from '@/views/agent-view'
import { ErrorsView } from '@/views/errors-view'
import { ErrorDetailView } from '@/views/error-detail-view'

type View =
  | { name: 'runs' }
  | { name: 'errors' }
  | { name: 'error-detail'; errorGroupId: string }
  | { name: 'flow'; runId: string }
  | { name: 'agent'; runId: string; agentId: string }

function App() {
  const [view, setView] = useState<View>({ name: 'runs' })
  const refreshRef = useRef<(() => void) | null>(null)

  const handleRefresh = useCallback(() => {
    refreshRef.current?.()
  }, [])

  const activeNav = view.name === 'errors' || view.name === 'error-detail' ? 'errors' : 'runs'

  const title =
    view.name === 'runs'
      ? 'Runs'
      : view.name === 'errors'
        ? 'Errors'
        : view.name === 'error-detail'
          ? 'Error Detail'
          : view.name === 'flow'
            ? 'Flow'
            : 'Detalle de agente'

  const handleNavigate = (navView: string) => {
    if (navView === 'errors') {
      setView({ name: 'errors' })
    } else {
      setView({ name: 'runs' })
    }
  }

  return (
    <AppShell
      title={title}
      activeView={activeNav}
      onNavigate={handleNavigate}
      onRefresh={handleRefresh}
    >
      <ErrorBoundary fallbackMessage="Error al renderizar la vista">
        {view.name === 'runs' && (
          <RunsView
            onRowClick={(id) => setView({ name: 'flow', runId: id })}
            refreshRef={refreshRef}
          />
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
        {view.name === 'errors' && (
          <ErrorsView
            onGroupClick={(id) => setView({ name: 'error-detail', errorGroupId: id })}
          />
        )}
        {view.name === 'error-detail' && (
          <ErrorDetailView
            groupId={view.errorGroupId}
            onBack={() => setView({ name: 'errors' })}
            onRunClick={(runId) => setView({ name: 'flow', runId })}
          />
        )}
      </ErrorBoundary>
    </AppShell>
  )
}

export default App
