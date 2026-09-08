import { useEffect, useMemo, useState } from 'react'
import { Pencil, Plus, Search, Trash2 } from 'lucide-react'
import {
  createRule,
  deleteRule,
  getRules,
  Rule,
  Algorithm,
  updateRule,
} from '../api/rules'

interface RulesPageProps {
  onCreate?: () => void
}

const emptyRule: Rule = {
  service: '',
  resource: '',
  algorithm: 'token_bucket',
  status: 'active',
  config: {
    capacity: 120,
    refill_rate: 2,
    key_strategy: 'account',
  },
}

export default function RulesPage() {
  const [rules, setRules] = useState<Rule[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [editing, setEditing] = useState<Rule | null>(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  async function loadRules() {
    try {
      setLoading(true)
      setRules(await getRules())
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to load rules',
      )
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadRules()
  }, [])

  const filteredRules = useMemo(() => {
    const value = search.toLowerCase().trim()

    if (!value) {
      return rules
    }

    return rules.filter(
      (rule) =>
        rule.service.toLowerCase().includes(value) ||
        rule.resource.toLowerCase().includes(value) ||
        rule.algorithm.toLowerCase().includes(value),
    )
  }, [rules, search])

  async function handleSave(rule: Rule) {
    try {
      setSaving(true)
      setError('')

      if (rule.id) {
        await updateRule(rule)
      } else {
        await createRule(rule)
      }

      setEditing(null)
      await loadRules()
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to save rule',
      )
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(rule: Rule) {
    const confirmed = window.confirm(
      `Delete ${rule.service}/${rule.resource}?`,
    )

    if (!confirmed) {
      return
    }

    try {
      await deleteRule(rule.service, rule.resource)
      await loadRules()
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'Failed to delete rule',
      )
    }
  }

  if (editing) {
    return (
      <RuleEditor
        rule={editing}
        saving={saving}
        onCancel={() => setEditing(null)}
        onSave={handleSave}
      />
    )
  }

  return (
    <div>
      <div className="page-header">
        <div>
          <p className="eyebrow">CONFIGURATION</p>
          <h1>Rules</h1>
          <p className="page-description">
            Define how requests are rate limited.
          </p>
        </div>

        <button
          className="primary-button"
          onClick={() => setEditing(emptyRule)}
        >
          <Plus size={16} />
          New rule
        </button>
      </div>

      {error && <div className="error-banner">{error}</div>}

      <div className="panel">
        <div className="table-toolbar">
          <div className="search-box">
            <Search size={16} />
            <input
              placeholder="Search services or resources..."
              value={search}
              onChange={(event) =>
                setSearch(event.target.value)
              }
            />
          </div>

          <span className="result-count">
            {filteredRules.length} rule
            {filteredRules.length === 1 ? '' : 's'}
          </span>
        </div>

        {loading ? (
          <div className="empty-state">
            Loading rules...
          </div>
        ) : filteredRules.length === 0 ? (
          <div className="empty-state">
            <GaugePlaceholder />

            <strong>
              {search ? 'No matching rules' : 'No rules yet'}
            </strong>

            <span>
              {search
                ? 'Try a different search.'
                : 'Create a rule to start protecting an API resource.'}
            </span>
          </div>
        ) : (
          <div className="table-wrapper">
            <table>
              <thead>
                <tr>
                  <th>Service</th>
                  <th>Resource</th>
                  <th>Algorithm</th>
                  <th>Configuration</th>
                  <th>Status</th>
                  <th />
                </tr>
              </thead>

              <tbody>
                {filteredRules.map((rule) => (
                  <tr
                    key={`${rule.service}:${rule.resource}`}
                  >
                    <td>
                      <strong>{rule.service}</strong>
                    </td>

                    <td className="mono">
                      {rule.resource}
                    </td>

                    <td>
                      <AlgorithmBadge
                        algorithm={rule.algorithm}
                      />
                    </td>

                    <td>
                      <RuleConfigSummary rule={rule} />
                    </td>

                    <td>
                      <span
                        className={`status-badge ${rule.status}`}
                      >
                        {rule.status}
                      </span>
                    </td>

                    <td>
                      <div className="row-actions">
                        <button
                          className="icon-button"
                          title="Edit rule"
                          onClick={() => setEditing(rule)}
                        >
                          <Pencil size={15} />
                        </button>

                        <button
                          className="icon-button danger"
                          title="Delete rule"
                          onClick={() =>
                            handleDelete(rule)
                          }
                        >
                          <Trash2 size={15} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

function RuleEditor({
  rule,
  saving,
  onCancel,
  onSave,
}: {
  rule: Rule
  saving: boolean
  onCancel: () => void
  onSave: (rule: Rule) => void
}) {
    const [draft, setDraft] = useState<Rule>(rule)
  const isTokenBucket = rule.algorithm === 'token_bucket'

  function updateAlgorithm(algorithm: Algorithm) {
    if (algorithm === 'token_bucket') {
      onSavePreview({
        ...draft,
        algorithm,
        config: {
          capacity: 120,
          refill_rate: 2,
          key_strategy: 'account',
        },
      })
    } else {
      onSavePreview({
        ...draft,
        algorithm,
        config: {
          limit: 5,
          window_seconds: 10,
          key_strategy: 'account',
        },
      })
    }
  }

  function onSavePreview(nextRule: Rule) {
    setDraft(nextRule)
  }

  return (
    <div>
      <div className="page-header">
        <div>
          <p className="eyebrow">
            {rule.id ? 'EDIT RULE' : 'NEW RULE'}
          </p>

          <h1>
            {rule.id ? 'Edit rule' : 'Create rule'}
          </h1>

          <p className="page-description">
            Configure the limits applied to this resource.
          </p>
        </div>
      </div>

      <div className="editor-layout">
        <section className="panel">
          <div className="form-section">
            <h2>Resource</h2>

            <div className="form-grid">
              <Field
                label="Service"
                value={draft.service}
                disabled={Boolean(rule.id)}
                placeholder="xoxoday"
                onChange={(value) =>
                  setDraft({
                    ...draft,
                    service: value,
                  } as Rule)
                }
              />

              <Field
                label="Resource"
                value={draft.resource}
                disabled={Boolean(rule.id)}
                placeholder="purchase"
                onChange={(value) =>
                  setDraft({
                    ...draft,
                    resource: value,
                  } as Rule)
                }
              />
            </div>
          </div>

          <div className="form-section">
            <h2>Algorithm</h2>

            <div className="algorithm-options">
              <AlgorithmOption
                selected={
                  draft.algorithm === 'token_bucket'
                }
                title="Token bucket"
                description="Allows bursts while controlling the average request rate."
                onClick={() =>
                  updateAlgorithm('token_bucket')
                }
              />

              <AlgorithmOption
                selected={
                  draft.algorithm === 'fixed_window'
                }
                title="Fixed window"
                description="Simple request count over fixed time windows."
                onClick={() =>
                  updateAlgorithm('fixed_window')
                }
              />

              <AlgorithmOption
                selected={
                  draft.algorithm === 'sliding_window'
                }
                title="Sliding window"
                description="Smooth request limits across a rolling time window."
                onClick={() =>
                  updateAlgorithm('sliding_window')
                }
              />
            </div>
          </div>

          <div className="form-section">
            <h2>Limit</h2>

            {isTokenBucket ? (
              <div className="form-grid">
                <NumberField
                  label="Capacity"
                  value={
                    'capacity' in draft.config
                      ? draft.config.capacity
                      : 120
                  }
                  onChange={(value) =>
                    setDraft({
                      ...draft,
                      config: {
                        capacity: value,
                        refill_rate:
                          'refill_rate' in draft.config
                            ? draft.config.refill_rate
                            : 2,
                        key_strategy: 'account',
                      },
                    } as Rule)
                  }
                />

                <NumberField
                  label="Refill rate / second"
                  value={
                    'refill_rate' in draft.config
                      ? draft.config.refill_rate
                      : 2
                  }
                  step={0.1}
                  onChange={(value) =>
                    setDraft({
                      ...draft,
                      config: {
                        capacity:
                          'capacity' in draft.config
                            ? draft.config.capacity
                            : 120,
                        refill_rate: value,
                        key_strategy: 'account',
                      },
                    } as Rule)
                  }
                />
              </div>
            ) : (
              <div className="form-grid">
                <NumberField
                  label="Request limit"
                  value={
                    'limit' in draft.config
                      ? draft.config.limit
                      : 5
                  }
                  onChange={(value) =>
                    setDraft({
                      ...draft,
                      config: {
                        limit: value,
                        window_seconds:
                          'window_seconds' in draft.config
                            ? draft.config.window_seconds
                            : 10,
                        key_strategy: 'account',
                      },
                    } as Rule)
                  }
                />

                <NumberField
                  label="Window / seconds"
                  value={
                    'window_seconds' in draft.config
                      ? draft.config.window_seconds
                      : 10
                  }
                  onChange={(value) =>
                    setDraft({
                      ...draft,
                      config: {
                        limit:
                          'limit' in draft.config
                            ? draft.config.limit
                            : 5,
                        window_seconds: value,
                        key_strategy: 'account',
                      },
                    } as Rule)
                  }
                />
              </div>
            )}

            <div className="field">
              <label>Key strategy</label>

              <select
                value={draft.config.key_strategy}
                onChange={(event) =>
                  setDraft({
                    ...draft,
                    config: {
                      ...draft.config,
                      key_strategy: event.target.value,
                    },
                  } as Rule)
                }
              >
                <option value="account">Account</option>
              </select>
            </div>
          </div>

          <div className="form-section">
            <h2>Status</h2>

            <select
              value={draft.status}
              onChange={(event) =>
                setDraft({
                  ...draft,
                  status: event.target.value as
                    | 'active'
                    | 'inactive',
                })
              }
            >
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>

          <div className="form-actions">
            <button
              className="secondary-button"
              onClick={onCancel}
              disabled={saving}
            >
              Cancel
            </button>

            <button
              className="primary-button"
              disabled={
                saving ||
                !draft.service.trim() ||
                !draft.resource.trim()
              }
              onClick={() => onSave(draft)}
            >
              {saving ? 'Saving...' : 'Save rule'}
            </button>
          </div>
        </section>

        <aside className="preview-panel">
          <span className="eyebrow">PREVIEW</span>

          <h3>{draft.service || 'service'}</h3>

          <p className="mono">
            {draft.resource || 'resource'}
          </p>

          <div className="preview-divider" />

          <div className="preview-row">
            <span>Algorithm</span>
            <strong>{algorithmLabel(draft.algorithm)}</strong>
          </div>

          {draft.algorithm === 'token_bucket' ? (
            <>
              <div className="preview-row">
                <span>Capacity</span>
                <strong>
                  {'capacity' in draft.config
                    ? draft.config.capacity
                    : 120}
                </strong>
              </div>

              <div className="preview-row">
                <span>Refill</span>
                <strong>
                  {'refill_rate' in draft.config
                    ? `${draft.config.refill_rate}/sec`
                    : '2/sec'}
                </strong>
              </div>
            </>
          ) : (
            <>
              <div className="preview-row">
                <span>Limit</span>
                <strong>
                  {'limit' in draft.config
                    ? draft.config.limit
                    : 5}
                </strong>
              </div>

              <div className="preview-row">
                <span>Window</span>
                <strong>
                  {'window_seconds' in draft.config
                    ? `${draft.config.window_seconds}s`
                    : '10s'}
                </strong>
              </div>
            </>
          )}

          <div className="preview-row">
            <span>Status</span>
            <span
              className={`status-badge ${draft.status}`}
            >
              {draft.status}
            </span>
          </div>
        </aside>
      </div>
    </div>
  )
}

function Field({
  label,
  value,
  placeholder,
  disabled,
  onChange,
}: {
  label: string
  value: string
  placeholder?: string
  disabled?: boolean
  onChange: (value: string) => void
}) {
  return (
    <div className="field">
      <label>{label}</label>

      <input
        value={value}
        placeholder={placeholder}
        disabled={disabled}
        onChange={(event) =>
          onChange(event.target.value)
        }
      />
    </div>
  )
}

function NumberField({
  label,
  value,
  step = 1,
  onChange,
}: {
  label: string
  value: number
  step?: number
  onChange: (value: number) => void
}) {
  return (
    <div className="field">
      <label>{label}</label>

      <input
        type="number"
        min="0"
        step={step}
        value={value}
        onChange={(event) =>
          onChange(Number(event.target.value))
        }
      />
    </div>
  )
}

function AlgorithmOption({
  selected,
  title,
  description,
  onClick,
}: {
  selected: boolean
  title: string
  description: string
  onClick: () => void
}) {
  return (
    <button
      className={`algorithm-option ${
        selected ? 'selected' : ''
      }`}
      onClick={onClick}
      type="button"
    >
      <div>
        <strong>{title}</strong>
        <span>{description}</span>
      </div>
    </button>
  )
}

function AlgorithmBadge({
  algorithm,
}: {
  algorithm: Algorithm
}) {
  return (
    <span className="algorithm-badge">
      {algorithmLabel(algorithm)}
    </span>
  )
}

function RuleConfigSummary({ rule }: { rule: Rule }) {
  if (rule.algorithm === 'token_bucket') {
    return (
      <span className="config-summary">
        {rule.config.capacity} capacity ·{' '}
        {rule.config.refill_rate}/sec
      </span>
    )
  }

  return (
    <span className="config-summary">
      {rule.config.limit} / {rule.config.window_seconds}s
    </span>
  )
}

function algorithmLabel(algorithm: Algorithm) {
  switch (algorithm) {
    case 'token_bucket':
      return 'Token bucket'
    case 'fixed_window':
      return 'Fixed window'
    case 'sliding_window':
      return 'Sliding window'
  }
}

function GaugePlaceholder() {
  return <div className="empty-icon">+</div>
}