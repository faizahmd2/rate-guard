import { useEffect, useState } from 'react'
import { Check, Copy, Plus, Trash2, X } from 'lucide-react'
import {
  createToken,
  getTokens,
  revokeToken,
  type APIToken,
} from '../api/tokens'

export default function TokensPage() {
  const [tokens, setTokens] = useState<APIToken[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [client, setClient] = useState('')
  const [creating, setCreating] = useState(false)

  const [createdToken, setCreatedToken] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const [revokingId, setRevokingId] = useState<string | null>(null)

  async function loadTokens() {
    try {
      setError('')
      const data = await getTokens()
      setTokens(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load tokens')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadTokens()
  }, [])

  async function handleCreate(event: React.FormEvent) {
    event.preventDefault()

    if (!name.trim() || !client.trim()) {
      return
    }

    try {
      setCreating(true)
      setError('')

      const result = await createToken(
        name.trim(),
        client.trim(),
      )

      setCreatedToken(result.token)

      setName('')
      setClient('')
      setShowCreate(false)

      await loadTokens()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create token')
    } finally {
      setCreating(false)
    }
  }

  async function handleRevoke(id: string) {
    const confirmed = window.confirm(
      'Are you sure you want to revoke this token? This cannot be undone.',
    )

    if (!confirmed) {
      return
    }

    try {
      setRevokingId(id)
      setError('')

      await revokeToken(id)

      setTokens((current) =>
        current.map((token) =>
          token.id === id
            ? { ...token, status: 'revoked' }
            : token,
        ),
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to revoke token')
    } finally {
      setRevokingId(null)
    }
  }

  async function handleCopy() {
    if (!createdToken) {
      return
    }

    await navigator.clipboard.writeText(createdToken)
    setCopied(true)

    setTimeout(() => {
      setCopied(false)
    }, 2000)
  }

  function formatDate(value: string) {
    return new Date(value).toLocaleString()
  }

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>API Tokens</h1>
          <p>
            Manage credentials used by applications to access RateGuard.
          </p>
        </div>

        <button
          className="primary-button"
          onClick={() => setShowCreate(true)}
        >
          <Plus size={16} />
          Create token
        </button>
      </div>

      {error && (
        <div className="error-banner">
          {error}
        </div>
      )}

      <div className="card">
        {loading ? (
          <div className="empty-state">
            Loading tokens...
          </div>
        ) : tokens.length === 0 ? (
          <div className="empty-state">
            <h3>No API tokens</h3>
            <p>
              Create a token to allow an application to access RateGuard.
            </p>
          </div>
        ) : (
          <div className="table-wrapper">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Client</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th></th>
                </tr>
              </thead>

              <tbody>
                {tokens.map((token) => (
                  <tr key={token.id}>
                    <td>
                      <div className="token-name">
                        {token.name}
                      </div>
                    </td>

                    <td>
                      <span className="muted">
                        {token.client}
                      </span>
                    </td>

                    <td>
                      <span
                        className={`status-badge ${
                          token.status === 'active'
                            ? 'status-active'
                            : 'status-revoked'
                        }`}
                      >
                        {token.status}
                      </span>
                    </td>

                    <td>
                      <span className="muted">
                        {formatDate(token.created_at)}
                      </span>
                    </td>

                    <td className="table-actions">
                      {token.status === 'active' && (
                        <button
                          className="icon-button danger"
                          title="Revoke token"
                          disabled={revokingId === token.id}
                          onClick={() => handleRevoke(token.id)}
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showCreate && (
        <div
          className="modal-backdrop"
          onClick={() => setShowCreate(false)}
        >
          <div
            className="modal"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="modal-header">
              <div>
                <h2>Create API token</h2>
                <p>
                  Create a credential for an application using RateGuard.
                </p>
              </div>

              <button
                className="icon-button"
                onClick={() => setShowCreate(false)}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleCreate}>
              <div className="form-group">
                <label htmlFor="token-name">
                  Token name
                </label>

                <input
                  id="token-name"
                  type="text"
                  placeholder="Production Xoxoday"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  autoFocus
                />
              </div>

              <div className="form-group">
                <label htmlFor="token-client">
                  Client
                </label>

                <input
                  id="token-client"
                  type="text"
                  placeholder="xoxoday-prod"
                  value={client}
                  onChange={(event) => setClient(event.target.value)}
                />
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  className="secondary-button"
                  onClick={() => setShowCreate(false)}
                >
                  Cancel
                </button>

                <button
                  type="submit"
                  className="primary-button"
                  disabled={
                    creating ||
                    !name.trim() ||
                    !client.trim()
                  }
                >
                  {creating ? 'Creating...' : 'Create token'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {createdToken && (
        <div className="modal-backdrop">
          <div className="modal token-created-modal">
            <div className="modal-header">
              <div>
                <h2>Token created</h2>
                <p>
                  Store this token securely. It won't be shown again.
                </p>
              </div>
            </div>

            <div className="token-display">
              <code>{createdToken}</code>

              <button
                className="icon-button"
                title="Copy token"
                onClick={handleCopy}
              >
                {copied ? (
                  <Check size={18} />
                ) : (
                  <Copy size={18} />
                )}
              </button>
            </div>

            {copied && (
              <div className="copy-success">
                Token copied to clipboard.
              </div>
            )}

            <div className="modal-actions">
              <button
                className="primary-button"
                onClick={() => {
                  setCreatedToken(null)
                  setCopied(false)
                }}
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}