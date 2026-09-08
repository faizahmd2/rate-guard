import { useEffect, useState } from 'react'
import { Database, Server } from 'lucide-react'
import { getSystemInfo, type SystemInfo } from '../api/system'

function Status({
  value,
}: {
  value: string
}) {
  const connected = value === 'connected'

  return (
    <span
      className={`settings-status ${
        connected
          ? 'settings-status-connected'
          : 'settings-status-disconnected'
      }`}
    >
      <span className="settings-status-dot" />
      {value}
    </span>
  )
}

export default function SettingsPage() {
  const [system, setSystem] = useState<SystemInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const data = await getSystemInfo()
        setSystem(data)
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : 'Failed to load system information',
        )
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [])

  if (loading) {
    return (
      <div className="page">
        <div className="page-header">
          <div>
            <h1>Settings</h1>
          </div>
        </div>

        <div className="card settings-simple-card">
          Loading...
        </div>
      </div>
    )
  }

  if (error || !system) {
    return (
      <div className="page">
        <div className="page-header">
          <div>
            <h1>Settings</h1>
          </div>
        </div>

        <div className="error-banner">
          {error || 'Unable to load system information'}
        </div>
      </div>
    )
  }

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>Settings</h1>
          <p>RateGuard instance information.</p>
        </div>
      </div>

      <div className="settings-simple-grid">
        <div className="card settings-simple-card">
          <div className="settings-simple-header">
            <Server size={18} />
            <h2>RateGuard</h2>
          </div>

          <div className="settings-simple-list">
            <div className="settings-simple-row">
              <span>Version</span>
              <strong>{system.version}</strong>
            </div>

            <div className="settings-simple-row">
              <span>Environment</span>
              <strong>{system.environment}</strong>
            </div>
          </div>
        </div>

        <div className="card settings-simple-card">
          <div className="settings-simple-header">
            <Database size={18} />
            <h2>Infrastructure</h2>
          </div>

          <div className="settings-simple-list">
            <div className="settings-simple-row">
              <span>Storage</span>
              <strong>{system.storage.driver}</strong>
            </div>

            <div className="settings-simple-row">
              <span>Redis</span>
              <Status value={system.redis.status} />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
