import { LogOut } from 'lucide-react'
import Sidebar, { Page } from './Sidebar'

interface LayoutProps {
  page: Page
  onNavigate: (page: Page) => void
  onLogout: () => void | Promise<void>
  children: React.ReactNode
}

export default function Layout({
  page,
  onNavigate,
  onLogout,
  children,
}: LayoutProps) {
  return (
    <div className="app-shell">
      <Sidebar
        page={page}
        onNavigate={onNavigate}
      />

      <div className="main-area">
        <header className="topbar">
          <div />

          <div className="topbar-user">
            <div className="avatar">A</div>

            <span>admin</span>

            <button
              className="icon-button"
              onClick={onLogout}
              title="Logout"
            >
              <LogOut size={17} />
            </button>
          </div>
        </header>

        <main className="content">
          {children}
        </main>
      </div>
    </div>
  )
}