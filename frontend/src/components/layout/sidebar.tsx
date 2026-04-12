import { Hammer, List } from 'lucide-react'
import { cn } from '@/lib/utils'

type NavView = 'runs'

interface SidebarProps {
  activeView: NavView
  onNavigate: (view: NavView) => void
}

interface NavItemConfig {
  view: NavView
  label: string
  icon: React.ReactNode
}

const navItems: NavItemConfig[] = [
  { view: 'runs', label: 'Runs', icon: <List size={16} /> },
]

export function Sidebar({ activeView, onNavigate }: SidebarProps) {
  return (
    <aside className="fixed left-0 top-0 h-full w-[220px] flex flex-col bg-card border-r border-border">
      {/* Logo */}
      <div className="flex items-center gap-2 px-4 py-4">
        <Hammer size={18} className="text-brand" />
        <span className="text-sm font-semibold tracking-wide text-foreground">anvil</span>
      </div>

      <div className="border-b border-border" />

      {/* Navegacion */}
      <nav className="flex-1 px-2 py-3 space-y-0.5">
        {navItems.map((item) => (
          <button
            key={item.view}
            onClick={() => onNavigate(item.view)}
            className={cn(
              'w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors text-left',
              activeView === item.view
                ? 'bg-[hsl(var(--brand-subtle))] text-[hsl(var(--brand-text))]'
                : 'text-muted-foreground hover:bg-secondary/60 hover:text-foreground',
            )}
          >
            {item.icon}
            {item.label}
          </button>
        ))}
      </nav>

      {/* Footer */}
      <div className="absolute bottom-0 left-0 right-0 px-4 py-3 border-t border-border">
        <div className="flex items-center gap-2">
          <span className="w-1.5 h-1.5 rounded-full bg-success flex-shrink-0" />
          <span className="font-mono text-[11px] text-muted-foreground">Watching</span>
        </div>
      </div>
    </aside>
  )
}
