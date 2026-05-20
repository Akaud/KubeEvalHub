import { useEffect, useState } from 'react'
import {
  FiPlus,
  FiCopy,
  FiCheck,
  FiX,
  FiRefreshCw,
} from 'react-icons/fi'
import { useAuth } from '../../context/AuthContext'
import { apiFetch } from '../../utils/apiFetch'

const DEFAULT_NAMESPACE = 'kubeevalhub-agent'
const DEFAULT_AGENT_IMAGE = 'docker.io/lewaldenko/kubeevalhubagent:1.0.4'
const getPublicBackendUrl = () => {
  if (typeof window === 'undefined') {
    return '/api'
  }
  return `${window.location.origin}/api`
}

export default function AgentsPage() {
  const [agents, setAgents] = useState([])

  const [isLoadingAgents, setIsLoadingAgents] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const [agentsError, setAgentsError] = useState('')

  const [showCreateAgentModal, setShowCreateAgentModal] = useState(false)
  const [agentName, setAgentName] = useState('')
  const [scrapeInterval, setScrapeInterval] = useState(30)
  const [isCreatingAgent, setIsCreatingAgent] = useState(false)
  const [createAgentError, setCreateAgentError] = useState('')
  const [createdAgentResult, setCreatedAgentResult] = useState(null)

  const [manifestCopied, setManifestCopied] = useState(false)
  const [togglingAgentId, setTogglingAgentId] = useState('')
  const [deletingAgentId, setDeletingAgentId] = useState('')

  const { isAuthenticated, isReady } = useAuth()

  const getInstallConfig = () => {
    return {
      backendUrl: getPublicBackendUrl(),
      image: DEFAULT_AGENT_IMAGE,
      namespace: DEFAULT_NAMESPACE,
    }
  }

  const buildInstallManifest = ({ token, scrapeIntervalSeconds }) => {
    const { backendUrl, image, namespace } = getInstallConfig()

    return `apiVersion: v1
kind: Namespace
metadata:
  name: ${namespace}

---
apiVersion: v1
kind: Secret
metadata:
  name: kubeevalhub-agent-secret
  namespace: ${namespace}
type: Opaque
stringData:
  AGENT_TOKEN: "${token}"

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubeevalhub-agent
  namespace: ${namespace}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeevalhub-agent
rules:
  - apiGroups: [""]
    resources: ["namespaces", "nodes", "pods"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "daemonsets", "replicasets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["metrics.k8s.io"]
    resources: ["nodes", "pods"]
    verbs: ["get", "list", "watch"]

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
    namespace: ${namespace}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubeevalhub-agent
  namespace: ${namespace}
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
          image: ${image}
          imagePullPolicy: IfNotPresent
          env:
            - name: BACKEND_URL
              value: "${backendUrl}"
            - name: AGENT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kubeevalhub-agent-secret
                  key: AGENT_TOKEN
            - name: SCRAPE_INTERVAL
              value: "${scrapeIntervalSeconds}s"
          resources:
            requests:
              cpu: "50m"
              memory: "64Mi"
            limits:
              cpu: "250m"
              memory: "256Mi"
`
  }

  const loadAgents = async ({ silent = false } = {}) => {
    if (!isAuthenticated) return

    if (!silent) {
      setIsLoadingAgents(true)
    }

    setAgentsError('')

    try {
      const response = await apiFetch('/api/agents')
      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to load agents.')
      }

      setAgents(Array.isArray(data) ? data : [])
    } catch (error) {
      setAgentsError(error.message || 'Failed to load agents.')
    } finally {
      if (!silent) {
        setIsLoadingAgents(false)
      }
    }
  }

  const refreshAll = async () => {
    if (!isAuthenticated) return

    setIsRefreshing(true)

    try {
      await loadAgents({ silent: true })
    } finally {
      setIsRefreshing(false)
    }
  }

  useEffect(() => {
    if (!isReady || !isAuthenticated) return
    loadAgents()
  }, [isReady, isAuthenticated])

  const openCreateAgentModal = () => {
    setAgentName('')
    setScrapeInterval(30)
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setManifestCopied(false)
    setShowCreateAgentModal(true)
  }

  const closeCreateAgentModal = () => {
    if (isCreatingAgent) return
    setShowCreateAgentModal(false)
    setAgentName('')
    setScrapeInterval(30)
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setManifestCopied(false)
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

    if (!isAuthenticated) {
      setCreateAgentError('Authentication is missing.')
      return
    }

    setIsCreatingAgent(true)
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setManifestCopied(false)

    try {
      const response = await apiFetch('/api/agents', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: trimmedName }),
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

  const handleDeleteAgent = async (agentId, agentNameValue) => {
    const confirmed = window.confirm(
      `Delete agent "${agentNameValue}"? This will stop backend authorization for that agent and remove its stored data.`
    )

    if (!confirmed) return

    setDeletingAgentId(agentId)
    setAgentsError('')

    try {
      const response = await apiFetch(`/api/agents/${agentId}`, {
        method: 'DELETE',
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to delete agent.')
      }

      if (createdAgentResult?.agent?.id === agentId) {
        setCreatedAgentResult(null)
        setManifestCopied(false)
      }

      await loadAgents({ silent: true })
    } catch (error) {
      setAgentsError(error.message || 'Failed to delete agent.')
    } finally {
      setDeletingAgentId('')
    }
  }

  const handleToggleAgent = async (agentId, enabled) => {
    setTogglingAgentId(agentId)
    setAgentsError('')

    try {
      const response = await apiFetch(`/api/agents/${agentId}/enabled`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
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

  const handleCopyManifest = async () => {
    const tokenValue = createdAgentResult?.token
    if (!tokenValue) return

    try {
      const manifest = buildInstallManifest({
        token: tokenValue,
        scrapeIntervalSeconds: scrapeInterval,
      })

      await navigator.clipboard.writeText(manifest)
      setManifestCopied(true)
      window.setTimeout(() => setManifestCopied(false), 2000)
    } catch (error) {
      setCreateAgentError(error.message || 'Failed to build YAML.')
      setManifestCopied(false)
    }
  }

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Agents</h1>
          <p>Create agents, copy installer YAML, and manage ingestion state.</p>
        </div>

        <div className="dashboard-header-actions">
          <button
            type="button"
            className="secondary-button"
            onClick={refreshAll}
            disabled={isRefreshing || isLoadingAgents}
          >
            <FiRefreshCw />
            <span>{isRefreshing ? 'Refreshing...' : 'Refresh'}</span>
          </button>

          <button
            className="primary-button"
            type="button"
            onClick={openCreateAgentModal}
          >
            <FiPlus />
            <span>Create agent</span>
          </button>
        </div>
      </div>

      <section className="dashboard-cards dashboard-cards-single">
        <div className="dashboard-card agents-panel">
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

                    <div className="cluster-meta-list">
                      <div className="cluster-meta-row">
                        <span>ID</span>
                        <strong>{agent.id}</strong>
                      </div>

                      <div className="cluster-meta-row">
                        <span>Enabled</span>
                        <strong>{agent.enabled ? 'Yes' : 'No'}</strong>
                      </div>

                      <div className="cluster-meta-row">
                        <span>Last heartbeat</span>
                        <strong>
                          {agent.lastHeartbeatAt
                            ? new Date(agent.lastHeartbeatAt).toLocaleString()
                            : 'Never'}
                        </strong>
                      </div>
                    </div>
                  </div>

                  <div className="agent-list-side">
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
                  <span>Scrape interval for installer (seconds)</span>
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
                  <button
                    type="button"
                    className="secondary-button"
                    onClick={closeCreateAgentModal}
                  >
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
                  <p>Attach the agent to a cluster and apply the YAML template below.</p>
                </div>

                <div className="result-meta">
                  <div>
                    <strong>Name:</strong> {createdAgentResult.agent?.name}
                  </div>
                  <div>
                    <strong>ID:</strong> {createdAgentResult.agent?.id}
                  </div>
                </div>

                {createAgentError && <div className="form-error">{createAgentError}</div>}

                <div className="copy-actions-grid">
                  <button
                    type="button"
                    className="secondary-button"
                    onClick={handleCopyManifest}
                  >
                    {manifestCopied ? <FiCheck /> : <FiCopy />}
                    <span>{manifestCopied ? 'Copied YAML' : 'Copy YAML'}</span>
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