import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '../../context/AuthContext'
import {
  FiPlus,
  FiCopy,
  FiCheck,
  FiX,
  FiRefreshCw,
} from 'react-icons/fi'

export default function AgentsPage() {
  const [showCreateAgentModal, setShowCreateAgentModal] = useState(false)
  const [agentName, setAgentName] = useState('')
  const [scrapeInterval, setScrapeInterval] = useState(30)
  const [isCreatingAgent, setIsCreatingAgent] = useState(false)
  const [createAgentError, setCreateAgentError] = useState('')
  const [createdAgentResult, setCreatedAgentResult] = useState(null)

  const [copied, setCopied] = useState(false)
  const [secretCommandCopied, setSecretCommandCopied] = useState(false)
  const [manifestCopied, setManifestCopied] = useState(false)
  const [installStepsCopied, setInstallStepsCopied] = useState(false)

  const [agents, setAgents] = useState([])
  const [isLoadingAgents, setIsLoadingAgents] = useState(false)
  const [isRefreshingAgents, setIsRefreshingAgents] = useState(false)
  const [agentsError, setAgentsError] = useState('')
  const [togglingAgentId, setTogglingAgentId] = useState('')
  const [deletingAgentId, setDeletingAgentId] = useState('')

  const { token } = useAuth()

  const authToken = useMemo(() => {
    return token || localStorage.getItem('token') || ''
  }, [token])

  const parseDurationToMs = (value) => {
    if (value == null) return null
    if (typeof value === 'number' && Number.isFinite(value)) return value

    const normalized = String(value).trim().toLowerCase()
    const match = normalized.match(/^(\d+(?:\.\d+)?)(ms|s|m|h)?$/)

    if (!match) return null

    const amount = Number(match[1])
    const unit = match[2] || 'ms'

    switch (unit) {
      case 'ms':
        return amount
      case 's':
        return amount * 1000
      case 'm':
        return amount * 60 * 1000
      case 'h':
        return amount * 60 * 60 * 1000
      default:
        return null
    }
  }

  const getEffectiveAgentStatus = (agent) => {
    if (!agent?.enabled) return 'disabled'
    if (!agent?.lastHeartbeatAt) return 'never_connected'

    const timeoutMs =
      parseDurationToMs(agent.responseTimeout) ??
      parseDurationToMs(agent.requestTimeout) ??
      parseDurationToMs(agent.timeout) ??
      10000

    const lastHeartbeatMs = new Date(agent.lastHeartbeatAt).getTime()

    if (Number.isNaN(lastHeartbeatMs)) {
      return agent.status || 'unknown'
    }

    const elapsedMs = Date.now() - lastHeartbeatMs

    if (elapsedMs > timeoutMs) {
      return 'failed'
    }

    return agent.status || 'online'
  }

  const buildSecretCommand = (tokenValue) => {
    return `kubectl create namespace kubeevalhub-agent
kubectl create secret generic kubeevalhub-agent-secret \\
  -n kubeevalhub-agent \\
  --from-literal=AGENT_TOKEN='${tokenValue}'`
  }

  const buildDeploymentManifest = (interval = '30s') => {
    return `apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubeevalhub-agent
  namespace: kubeevalhub-agent

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeevalhub-agent
rules:
  - apiGroups: [""]
    resources: ["namespaces", "nodes", "pods"]
    verbs: ["get", "list"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "daemonsets"]
    verbs: ["get", "list"]
  - apiGroups: ["metrics.k8s.io"]
    resources: ["nodes", "pods"]
    verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubeevalhub-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubeevalhub-agent
subjects:
  - kind: ServiceAccount
    name: kubeevalhub-agent
    namespace: kubeevalhub-agent

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubeevalhub-agent
  namespace: kubeevalhub-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubeevalhub-agent
  template:
    metadata:
      labels:
        app: kubeevalhub-agent
    spec:
      serviceAccountName: kubeevalhub-agent
      containers:
        - name: agent
          image: kubeevalhub-agent:latest
          imagePullPolicy: IfNotPresent
          env:
            - name: BACKEND_URL
              value: "http://host.minikube.internal:5000"
            - name: AGENT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kubeevalhub-agent-secret
                  key: AGENT_TOKEN
            - name: SCRAPE_INTERVAL
              value: "${interval}"
          resources:
            requests:
              cpu: "50m"
              memory: "64Mi"
            limits:
              cpu: "250m"
              memory: "256Mi"
`
  }

  const buildInstallSteps = (tokenValue) => {
    return `${buildSecretCommand(tokenValue)}

# save the deployment manifest to agent.yaml, then apply it
kubectl apply -f agent.yaml`
  }

  const loadAgents = async ({ silent = false } = {}) => {
    if (!authToken) return

    if (silent) {
      setIsRefreshingAgents(true)
    } else {
      setIsLoadingAgents(true)
    }

    setAgentsError('')

    try {
      const response = await fetch('/api/agents', {
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to load agents.')
      }

      setAgents(
        Array.isArray(data)
          ? data.map((agent) => ({
              ...agent,
              effectiveStatus: getEffectiveAgentStatus(agent),
            }))
          : []
      )
    } catch (error) {
      setAgentsError(error.message || 'Failed to load agents.')
    } finally {
      if (silent) {
        setIsRefreshingAgents(false)
      } else {
        setIsLoadingAgents(false)
      }
    }
  }

  useEffect(() => {
    if (!authToken) return

    loadAgents()

    const intervalId = window.setInterval(() => {
      loadAgents({ silent: true })
    }, 15000)

    return () => window.clearInterval(intervalId)
  }, [authToken])

  const openCreateAgentModal = () => {
    setAgentName('')
    setScrapeInterval(30)
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setCopied(false)
    setSecretCommandCopied(false)
    setManifestCopied(false)
    setInstallStepsCopied(false)
    setShowCreateAgentModal(true)
  }

  const closeCreateAgentModal = () => {
    if (isCreatingAgent) return
    setShowCreateAgentModal(false)
    setAgentName('')
    setScrapeInterval(30)
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setCopied(false)
    setSecretCommandCopied(false)
    setManifestCopied(false)
    setInstallStepsCopied(false)
  }

  const handleCreateAgent = async (e) => {
    e.preventDefault()

    const trimmedName = agentName.trim()

    if (!trimmedName) {
      setCreateAgentError('Agent name is required.')
      return
    }

    if (scrapeInterval < 5) {
      setCreateAgentError('Scrape interval must be at least 5 seconds.')
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
    setSecretCommandCopied(false)
    setManifestCopied(false)
    setInstallStepsCopied(false)

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
      await loadAgents({ silent: true })
    } catch (error) {
      setCreateAgentError(error.message || 'Failed to create agent.')
    } finally {
      setIsCreatingAgent(false)
    }
  }

  const handleDeleteAgent = async (agentId, agentName) => {
    if (!authToken) return

    const confirmed = window.confirm(
      `Delete agent "${agentName}"? This will stop backend authorization for that agent and remove its stored data.`
    )

    if (!confirmed) return

    setDeletingAgentId(agentId)
    setAgentsError('')

    try {
      const response = await fetch(`/api/agents/${agentId}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to delete agent.')
      }

      if (createdAgentResult?.agent?.id === agentId) {
        setCreatedAgentResult(null)
        setCopied(false)
        setSecretCommandCopied(false)
        setManifestCopied(false)
        setInstallStepsCopied(false)
      }

      await loadAgents({ silent: true })
    } catch (error) {
      setAgentsError(error.message || 'Failed to delete agent.')
    } finally {
      setDeletingAgentId('')
    }
  }

  const handleToggleAgent = async (agentId, enabled) => {
    if (!authToken) return

    setTogglingAgentId(agentId)
    setAgentsError('')

    try {
      const response = await fetch(`/api/agents/${agentId}/enabled`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authToken}`,
        },
        body: JSON.stringify({ enabled }),
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to update agent.')
      }

      await loadAgents({ silent: true })
    } catch (error) {
      setAgentsError(error.message || 'Failed to update agent.')
    } finally {
      setTogglingAgentId('')
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

  const handleCopySecretCommand = async () => {
    const tokenValue = createdAgentResult?.token
    if (!tokenValue) return

    try {
      await navigator.clipboard.writeText(buildSecretCommand(tokenValue))
      setSecretCommandCopied(true)
      window.setTimeout(() => setSecretCommandCopied(false), 2000)
    } catch {
      setSecretCommandCopied(false)
    }
  }

  const handleCopyManifest = async () => {
    try {
      await navigator.clipboard.writeText(buildDeploymentManifest(`${scrapeInterval}s`))
      setManifestCopied(true)
      window.setTimeout(() => setManifestCopied(false), 2000)
    } catch {
      setManifestCopied(false)
    }
  }

  const handleCopyInstallSteps = async () => {
    const tokenValue = createdAgentResult?.token
    if (!tokenValue) return

    try {
      await navigator.clipboard.writeText(buildInstallSteps(tokenValue))
      setInstallStepsCopied(true)
      window.setTimeout(() => setInstallStepsCopied(false), 2000)
    } catch {
      setInstallStepsCopied(false)
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

            <button
              className="create-agent-button create-agent-icon-button"
              type="button"
              onClick={openCreateAgentModal}
              aria-label="Create agent"
              title="Create agent"
            >
              <FiPlus />
            </button>
          </div>
        </div>
      </section>

      <section className="dashboard-cards dashboard-cards-single">
        <div className="dashboard-card agents-panel">
          <div className="agents-panel-header">
            <div>
              <h3>Existing agents</h3>
              <p>View current state and enable or disable ingestion.</p>
            </div>

            <button
              type="button"
              className="secondary-button"
              onClick={() => loadAgents({ silent: true })}
              disabled={isRefreshingAgents || isLoadingAgents}
            >
              <FiRefreshCw />
              <span>{isRefreshingAgents ? 'Refreshing...' : 'Refresh'}</span>
            </button>
          </div>

          {agentsError && <div className="form-error">{agentsError}</div>}

          {isLoadingAgents ? (
            <p>Loading agents...</p>
          ) : agents.length === 0 ? (
            <p>No agents created yet.</p>
          ) : (
            <div className="agents-list">
              {agents.map((agent) => (
                <div key={agent.id} className="agent-list-item">
                  <div className="agent-list-main">
                    <h4>{agent.name}</h4>
                    <p>ID: {agent.id}</p>
                    <p>Enabled: {String(agent.enabled)}</p>
                    <p>
                      Last heartbeat:{' '}
                      {agent.lastHeartbeatAt
                        ? new Date(agent.lastHeartbeatAt).toLocaleString()
                        : 'Never'}
                    </p>
                  </div>

                  <div className="agent-list-side">
                    <span className={`agent-status-badge status-${agent.effectiveStatus}`}>
                      {agent.effectiveStatus}
                    </span>

                    <button
                      type="button"
                      className={agent.enabled ? 'secondary-button' : 'primary-button'}
                      onClick={() => handleToggleAgent(agent.id, !agent.enabled)}
                      disabled={togglingAgentId === agent.id || deletingAgentId === agent.id}
                    >
                      {togglingAgentId === agent.id
                        ? 'Updating...'
                        : agent.enabled
                          ? 'Disable'
                          : 'Enable'}
                    </button>

                    <button
                      type="button"
                      className="secondary-button"
                      onClick={() => handleDeleteAgent(agent.id, agent.name)}
                      disabled={deletingAgentId === agent.id || togglingAgentId === agent.id}
                    >
                      {deletingAgentId === agent.id ? 'Deleting...' : 'Delete'}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
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

                <label className="field-group">
                  <span>Scrape interval for template (seconds)</span>
                  <input
                    type="number"
                    min={5}
                    step={1}
                    value={scrapeInterval}
                    onChange={(e) => setScrapeInterval(Number(e.target.value))}
                    placeholder="30"
                    disabled={isCreatingAgent}
                  />
                </label>

                <div className="agent-default-state">
                  <div className="state-badge state-badge-disabled">minimum: 5s</div>
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
                  <p>Copy the token and installation assets now. The token is shown only once.</p>
                </div>

                <div className="result-meta">
                  <div>
                    <strong>Name:</strong> {createdAgentResult.agent?.name}
                  </div>
                  <div>
                    <strong>ID:</strong> {createdAgentResult.agent?.id}
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

                <div className="copy-actions-grid">
                  <button type="button" className="secondary-button" onClick={handleCopySecretCommand}>
                    {secretCommandCopied ? <FiCheck /> : <FiCopy />}
                    <span>{secretCommandCopied ? 'Copied secret command' : 'Copy secret command'}</span>
                  </button>

                  <button type="button" className="secondary-button" onClick={handleCopyManifest}>
                    {manifestCopied ? <FiCheck /> : <FiCopy />}
                    <span>{manifestCopied ? 'Copied agent template' : 'Copy agent template'}</span>
                  </button>

                  <button type="button" className="secondary-button" onClick={handleCopyInstallSteps}>
                    {installStepsCopied ? <FiCheck /> : <FiCopy />}
                    <span>{installStepsCopied ? 'Copied install steps' : 'Copy install steps'}</span>
                  </button>
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