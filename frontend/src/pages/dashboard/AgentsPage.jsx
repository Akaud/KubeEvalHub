import { useMemo, useState } from 'react'
import { useAuth } from '../../context/AuthContext'
import {
  FiPlus,
  FiCopy,
  FiCheck,
  FiX,
} from 'react-icons/fi'

export default function AgentsPage() {
  const [showCreateAgentModal, setShowCreateAgentModal] = useState(false)
  const [agentName, setAgentName] = useState('')
  const [isCreatingAgent, setIsCreatingAgent] = useState(false)
  const [createAgentError, setCreateAgentError] = useState('')
  const [createdAgentResult, setCreatedAgentResult] = useState(null)
  const [copied, setCopied] = useState(false)

  const { token } = useAuth()

  const authToken = useMemo(() => {
    return token || localStorage.getItem('token') || ''
  }, [token])

  const openCreateAgentModal = () => {
    setAgentName('')
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setCopied(false)
    setShowCreateAgentModal(true)
  }

  const closeCreateAgentModal = () => {
    if (isCreatingAgent) return
    setShowCreateAgentModal(false)
    setAgentName('')
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setCopied(false)
  }

  const handleCreateAgent = async (e) => {
    e.preventDefault()

    const trimmedName = agentName.trim()

    if (!trimmedName) {
      setCreateAgentError('Agent name is required.')
      return
    }

    if (!authToken) {
      setCreateAgentError('Authentication token is missing.')
      return
    }

    setIsCreatingAgent(true)
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setCopied(false)

    try {
      const response = await fetch('/api/agents', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authToken}`,
        },
        body: JSON.stringify({ name: trimmedName }),
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to create agent.')
      }

      setCreatedAgentResult(data)
      setAgentName('')
    } catch (error) {
      setCreateAgentError(error.message || 'Failed to create agent.')
    } finally {
      setIsCreatingAgent(false)
    }
  }

  const handleCopyToken = async () => {
    const tokenToCopy = createdAgentResult?.token
    if (!tokenToCopy) return

    try {
      await navigator.clipboard.writeText(tokenToCopy)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      setCopied(false)
    }
  }

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Agents</h1>
          <p>Create and manage agents connected to your self-hosted cluster.</p>
        </div>
      </div>

      <section className="dashboard-cards dashboard-cards-single">
        <div className="dashboard-card dashboard-card-primary agents-panel">
          <div className="agents-panel-header">
            <div>
              <h3>Agents</h3>
              <p>Create a new agent and get its connection token.</p>
            </div>

            <button className="create-agent-button" type="button" onClick={openCreateAgentModal}>
              <FiPlus />
              <span>Create agent</span>
            </button>
          </div>
        </div>
      </section>

      {showCreateAgentModal && (
        <div className="modal-backdrop" onClick={closeCreateAgentModal}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h3>Create agent</h3>
                <p>Create a cluster agent and retrieve its connection token.</p>
              </div>

              <button className="icon-button" type="button" onClick={closeCreateAgentModal}>
                <FiX />
              </button>
            </div>

            {!createdAgentResult ? (
              <form onSubmit={handleCreateAgent} className="agent-form">
                <label className="field-group">
                  <span>Agent name</span>
                  <input
                    type="text"
                    value={agentName}
                    onChange={(e) => setAgentName(e.target.value)}
                    placeholder="Enter agent name"
                    maxLength={128}
                    disabled={isCreatingAgent}
                  />
                </label>

                <div className="agent-default-state">
                  <div className="state-badge state-badge-disabled">enabled: false</div>
                  <div className="state-badge state-badge-inactive">active: false</div>
                </div>

                {createAgentError && <div className="form-error">{createAgentError}</div>}

                <div className="modal-actions">
                  <button type="button" className="secondary-button" onClick={closeCreateAgentModal}>
                    Cancel
                  </button>

                  <button type="submit" className="primary-button" disabled={isCreatingAgent}>
                    {isCreatingAgent ? 'Creating...' : 'Create agent'}
                  </button>
                </div>
              </form>
            ) : (
              <div className="created-agent-result">
                <div className="success-box">
                  <h4>Agent created</h4>
                  <p>Copy this token now. It should be shown only once.</p>
                </div>

                <div className="result-meta">
                  <div>
                    <strong>Name:</strong> {createdAgentResult.agent?.name}
                  </div>
                  <div>
                    <strong>ID:</strong> {createdAgentResult.agent?.id}
                  </div>
                  <div>
                    <strong>Enabled:</strong> {String(createdAgentResult.agent?.enabled ?? false)}
                  </div>
                  <div>
                    <strong>Active:</strong> {String(createdAgentResult.agent?.active ?? false)}
                  </div>
                </div>

                <div className="token-block">
                  <label>Agent token</label>
                  <div className="token-row">
                    <code>{createdAgentResult.token}</code>
                    <button type="button" className="copy-button" onClick={handleCopyToken}>
                      {copied ? <FiCheck /> : <FiCopy />}
                      <span>{copied ? 'Copied' : 'Copy'}</span>
                    </button>
                  </div>
                </div>

                <div className="modal-actions">
                  <button type="button" className="primary-button" onClick={closeCreateAgentModal}>
                    Done
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </>
  )
}