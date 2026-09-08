import { useEffect, useState } from 'react'
import {
  ArrowRight,
  Database,
  Gauge,
  KeyRound,
  Plus,
  Server,
} from 'lucide-react'
import { getRules, Rule } from '../api/rules'
import { getTokens } from '../api/tokens'
import { getSystemInfo, SystemInfo } from '../api/system'

interface DashboardPageProps {
  onNavigate: (page: 'rules' | 'tokens') => void
}

export default function DashboardPage({
  onNavigate,
}: DashboardPageProps) {
  const [rules, setRules] = useState<Rule[]>([])
  const [tokensCount, setTokensCount] = useState(0)
  const [system, setSystem] = useState<SystemInfo | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function loadDashboard() {
      try {
        const [rulesData, tokensData, systemData] =
          await Promise.all([
            getRules(),
            getTokens(),
            getSystemInfo(),
          ])

        setRules(rulesData)
        setTokensCount(
          tokensData.filter(
            (token) => token.status === 'active',
          ).length,
        )
        setSystem(systemData)
      } catch {
        // Keep the dashboard usable even if one request fails.
      } finally {
        setLoading(false)
      }
    }

    loadDashboard()
  }, [])

  const activeRules = rules.filter(
    (rule) => rule.status === 'active',
  )

  const systemHealthy =
    system?.storage.status === 'connected' &&
    system?.redis.status === 'connected'

  return (
    <div>
      <div className="page-header">
        <div>
          <p className="eyebrow">OVERVIEW</p>

          <h1>Dashboard</h1>

          <p className="page-description">
            Manage your rate limiting configuration and API access.
          </p>
        </div>

        <button
          className="primary-button"
          onClick={() => onNavigate('rules')}
        >
          <Plus size={16} />
          New rule
        </button>
      </div>

      <div className="stats-grid">
        <StatCard
          icon={<Gauge size={18} />}
          label="Total rules"
          value={loading ? '—' : rules.length}
        />

        <StatCard
          icon={<ActivityIcon />}
          label="Active rules"
          value={loading ? '—' : activeRules.length}
        />

        <StatCard
          icon={<KeyRound size={18} />}
          label="API tokens"
          value={loading ? '—' : tokensCount}
        />

        <StatCard
          icon={<Server size={18} />}
          label="System"
          value={
            loading
              ? '—'
              : systemHealthy
                ? 'Healthy'
                : 'Attention'
          }
          positive={!loading && systemHealthy}
        />
      </div>

      <div className="section-grid">
        <section className="panel">
          <div className="panel-header">
            <div>
              <h2>Rate limiting</h2>
              <p>Current rule configuration.</p>
            </div>

            <button
              className="text-button"
              onClick={() => onNavigate('rules')}
            >
              View rules
              <ArrowRight size={15} />
            </button>
          </div>

          {loading ? (
            <div className="empty-state">
              Loading rules...
            </div>
          ) : rules.length === 0 ? (
            <div className="empty-state">
              <Gauge size={24} />

              <strong>No rules configured</strong>

              <span>
                Create your first rate limit rule to get started.
              </span>

              <button
                className="primary-button"
                onClick={() => onNavigate('rules')}
              >
                Create rule
              </button>
            </div>
          ) : (
            <div className="rule-summary">
              {rules.slice(0, 5).map((rule) => (
                <div
                  className="rule-summary-row"
                  key={`${rule.service}:${rule.resource}`}
                >
                  <div>
                    <strong>{rule.service}</strong>
                    <span>{rule.resource}</span>
                  </div>

                  <span
                    className={`status-badge ${rule.status}`}
                  >
                    {rule.status}
                  </span>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="panel">
          <div className="panel-header">
            <div>
              <h2>System</h2>
              <p>RateGuard dependencies.</p>
            </div>
          </div>

          <div className="health-list">
            <HealthRow
              icon={<Server size={17} />}
              label="RateGuard"
              status={loading ? 'Checking...' : 'Operational'}
            />

            <HealthRow
              icon={<Database size={17} />}
              label="Storage"
              status={
                loading
                  ? 'Checking...'
                  : formatStatus(system?.storage.status)
              }
            />

            <HealthRow
              icon={<Server size={17} />}
              label="Redis"
              status={
                loading
                  ? 'Checking...'
                  : formatStatus(system?.redis.status)
              }
            />
          </div>
        </section>
      </div>
    </div>
  )
}

function formatStatus(status?: string) {
  if (!status) {
    return 'Unavailable'
  }

  return status.charAt(0).toUpperCase() + status.slice(1)
}

function StatCard({
  icon,
  label,
  value,
  positive,
}: {
  icon: React.ReactNode
  label: string
  value: string | number
  positive?: boolean
}) {
  return (
    <div className="stat-card">
      <div className="stat-icon">{icon}</div>

      <span>{label}</span>

      <strong className={positive ? 'positive' : ''}>
        {value}
      </strong>
    </div>
  )
}

function HealthRow({
  icon,
  label,
  status,
}: {
  icon: React.ReactNode
  label: string
  status: string
}) {
  const healthy =
    status === 'Operational' ||
    status === 'Connected'

  return (
    <div className="health-row">
      <div className="health-label">
        {icon}
        <span>{label}</span>
      </div>

      <div className="health-status">
        <span
          className={`status-dot ${
            healthy ? '' : 'status-dot-warning'
          }`}
        />
        {status}
      </div>
    </div>
  )
}

function ActivityIcon() {
  return <Gauge size={18} />
}