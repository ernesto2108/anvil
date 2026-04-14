import { Sidebar } from './sidebar'
import { Topbar } from './topbar'

type NavView = 'runs'

interface AppShellProps {
  title: string
  activeView: NavView
  onNavigate: (view: NavView) => void
  actions?: React.ReactNode
  onRefresh?: () => void
  children: React.ReactNode
}

export function AppShell({ title, activeView, onNavigate, actions, onRefresh, children }: AppShellProps) {
  return (
    <div className="flex h-full bg-background">
      <Sidebar activeView={activeView} onNavigate={onNavigate} />
      <div className="flex-1 flex flex-col ml-[220px] min-h-0 min-w-0">
        <Topbar title={title} actions={actions} onRefresh={onRefresh} />
        <main className="flex-1 overflow-y-auto overflow-x-hidden flex flex-col min-h-0 min-w-0">{children}</main>
      </div>
    </div>
  )
}
