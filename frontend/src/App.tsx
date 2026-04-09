import { useState } from 'react'
import { AppShell } from '@/components/layout/app-shell'
import { RunsView } from '@/views/runs-view'
import { FlowView } from '@/views/flow-view'

type View = { name: 'runs' } | { name: 'flow'; runId: string }

function App() {
  const [view, setView] = useState<View>({ name: 'runs' })

  const activeNav = view.name === 'flow' ? 'runs' : (view.name as 'runs' | 'metrics')

  return (
    <AppShell
      title={view.name === 'runs' ? 'Runs' : 'Flow'}
      activeView={activeNav}
      onNavigate={(v) => setView({ name: v as 'runs' })}
    >
      {view.name === 'runs' && (
        <RunsView onRowClick={(id) => setView({ name: 'flow', runId: id })} />
      )}
      {/* La navegación al detalle del agente se implementa en DASH-FEAT-007. */}
      {view.name === 'flow' && (
        <FlowView
          runId={view.runId}
          onBack={() => setView({ name: 'runs' })}
        />
      )}
    </AppShell>
  )
}

export default App
