import { useEffect, useState } from 'react'
import { getAuthStatus, logout } from './api/auth'
import Layout from './components/Layout'
import { Page } from './components/Sidebar'
import DashboardPage from './pages/DashboardPage'
import LoginPage from './pages/LoginPage'
import RulesPage from './pages/RulesPage'
import SettingsPage from './pages/SettingsPage'
import SetupPage from './pages/SetupPage'
import TokensPage from './pages/TokensPage'

type AppState =
  | 'loading'
  | 'setup'
  | 'login'
  | 'authenticated'

export default function App() {
  const [state, setState] =
    useState<AppState>('loading')

  const [page, setPage] =
    useState<Page>('overview')

  useEffect(() => {
    checkAuth()
  }, [])

  async function checkAuth() {
    try {
      const status = await getAuthStatus()

      if (status.setup_required) {
        setState('setup')
      } else if (status.authenticated) {
        setState('authenticated')
      } else {
        setState('login')
      }
    } catch {
      setState('login')
    }
  }

  if (state === 'loading') {
    return (
      <div className="loading">
        Loading RateGuard...
      </div>
    )
  }

  if (state === 'setup') {
    return (
      <SetupPage
        onComplete={() => {
          setState('authenticated')
          setPage('overview')
        }}
      />
    )
  }

  if (state === 'login') {
    return (
      <LoginPage
        onLogin={() => {
          setState('authenticated')
          setPage('overview')
        }}
      />
    )
  }

  function renderPage() {
    switch (page) {
      case 'overview':
        return (
          <DashboardPage
            onNavigate={(nextPage) =>
              setPage(nextPage)
            }
          />
        )

      case 'rules':
        return <RulesPage />

      case 'tokens':
        return <TokensPage />

      case 'settings':
        return <SettingsPage />
    }
  }

  return (
    <Layout
      page={page}
      onNavigate={setPage}
      onLogout={async () => {
        await logout()
        setState('login')
        setPage('overview')
      }}
    >
      {renderPage()}
    </Layout>
  )
}