import { Sidebar } from './sidebar'
import { Topbar } from './topbar'

type NavView = 'runs' | 'flow' | 'metrics'

interface AppShellProps {
  title: string
  activeView: NavView
  onNavigate: (view: NavView) => void
  children: React.ReactNode
}

export function AppShell({ title, activeView, onNavigate, children }: AppShellProps) {
  return (
    <div className="flex h-full bg-background">
      <Sidebar activeView={activeView} onNavigate={onNavigate} />
      <div className="flex-1 flex flex-col ml-[220px] min-h-0">
        <Topbar title={title} />
        <main className="flex-1 overflow-auto">{children}</main>
      </div>
    </div>
  )
}
