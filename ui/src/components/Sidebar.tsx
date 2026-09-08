import {
  Activity,
  Gauge,
  KeyRound,
  Settings,
} from 'lucide-react'

export type Page =
  | 'overview'
  | 'rules'
  | 'tokens'
  | 'settings'

interface SidebarProps {
  page: Page
  onNavigate: (page: Page) => void
}

export default function Sidebar({
  page,
  onNavigate,
}: SidebarProps) {
  const items = [
    {
      id: 'overview' as const,
      label: 'Overview',
      icon: Activity,
    },
    {
      id: 'rules' as const,
      label: 'Rules',
      icon: Gauge,
    },
    {
      id: 'tokens' as const,
      label: 'API Tokens',
      icon: KeyRound,
    },
    {
      id: 'settings' as const,
      label: 'Settings',
      icon: Settings,
    },
  ]

  return (
    <aside className="sidebar">
      <div className="sidebar-brand">
        <div className="brand-mark">R</div>

        <div>
          <strong>RateGuard</strong>
          <span>Rate limiting</span>
        </div>
      </div>

      <nav className="sidebar-nav">
        {items.map((item) => {
          const Icon = item.icon

          return (
            <button
              key={item.id}
              className={`nav-item ${
                page === item.id ? 'active' : ''
              }`}
              onClick={() => onNavigate(item.id)}
            >
              <Icon size={17} />
              <span>{item.label}</span>
            </button>
          )
        })}
      </nav>

      <div className="sidebar-footer">
        <div className="system-indicator">
          <span className="status-dot" />
          <span>System operational</span>
        </div>

        <span className="version">RateGuard v0.1</span>
      </div>
    </aside>
  )
}