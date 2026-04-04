import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { useNavigate } from 'react-router-dom'
import {
  FiUser,
  FiServer,
  FiSettings,
  FiLogOut,
  FiSun,
  FiMoon,
  FiPlus,
  FiCopy,
  FiCheck,
  FiX,
} from 'react-icons/fi'
import '../styles/DashboardPage.css'

export default function DashboardPage() {
  const [activeSection, setActiveSection] = useState('profile')
  const [theme, setTheme] = useState(localStorage.getItem('theme') || 'light')

  const [showCreateAgentModal, setShowCreateAgentModal] = useState(false)
  const [agentName, setAgentName] = useState('')
  const [isCreatingAgent, setIsCreatingAgent] = useState(false)
  const [createAgentError, setCreateAgentError] = useState('')
  const [createdAgentResult, setCreatedAgentResult] = useState(null)
  const [copied, setCopied] = useState(false)

  const { logout, token } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    document.body.classList.remove('theme-light', 'theme-dark')
    document.body.classList.add(theme === 'dark' ? 'theme-dark' : 'theme-light')
    localStorage.setItem('theme', theme)
  }, [theme])

  const authToken = useMemo(() => {
    return token || localStorage.getItem('token') || ''
  }, [token])

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

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
        body: JSON.stringify({
          name: trimmedName,
        }),
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
    <div className={`dashboard-layout ${theme === 'dark' ? 'dashboard-layout-dark' : ''}`}>
      <aside className="dashboard-sidebar">
        <div className="sidebar-brand">
          <h2>Donezo</h2>
          <p>Toolbox</p>
        </div>

        <nav className="sidebar-nav">
          <button
            className={`sidebar-item ${activeSection === 'profile' ? 'active' : ''}`}
            type="button"
            onClick={() => setActiveSection('profile')}
          >
            <FiUser />
            <span>Profile</span>
          </button>

          <button
            className={`sidebar-item ${activeSection === 'cluster' ? 'active' : ''}`}
            type="button"
            onClick={() => setActiveSection('cluster')}
          >
            <FiServer />
            <span>Cluster</span>
          </button>

          <button
            className={`sidebar-item ${activeSection === 'settings' ? 'active' : ''}`}
            type="button"
            onClick={() => setActiveSection('settings')}
          >
            <FiSettings />
            <span>Settings</span>
          </button>
        </nav>

        <button className="sidebar-logout" type="button" onClick={handleLogout}>
          <FiLogOut />
          <span>Logout</span>
        </button>
      </aside>

      <main className="dashboard-content">
        {activeSection === 'profile' && (
          <>
            <div className="dashboard-header">
              <div>
                <h1>Profile</h1>
                <p>View and manage your account information.</p>
              </div>
            </div>

            <section className="dashboard-cards">
              <div className="dashboard-card dashboard-card-primary">
                <h3>Profile</h3>
                <p>View and manage your account information.</p>
              </div>
              <div className="dashboard-card">
                <h3>Account</h3>
                <p>Inspect personal details and authentication data.</p>
              </div>
              <div className="dashboard-card">
                <h3>Security</h3>
                <p>Review password and access-related settings.</p>
              </div>
            </section>
          </>
        )}

        {activeSection === 'cluster' && (
          <>
            <div className="dashboard-header">
              <div>
                <h1>Cluster</h1>
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
          </>
        )}

        {activeSection === 'settings' && (
          <section className="settings-panel">
            <div className="settings-card">
              <h3>Appearance</h3>
              <p>Choose how the dashboard should look.</p>

              <div className="theme-toggle">
                <button
                  type="button"
                  className={`theme-button ${theme === 'light' ? 'active' : ''}`}
                  onClick={() => setTheme('light')}
                >
                  <FiSun />
                  <span>Light mode</span>
                </button>

                <button
                  type="button"
                  className={`theme-button ${theme === 'dark' ? 'active' : ''}`}
                  onClick={() => setTheme('dark')}
                >
                  <FiMoon />
                  <span>Dark mode</span>
                </button>
              </div>
            </div>
          </section>
        )}
      </main>

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
    </div>
  )
}