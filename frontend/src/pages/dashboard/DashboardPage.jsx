import { useEffect, useState } from 'react'
import {
  FiPlus,
  FiCopy,
  FiCheck,
  FiX,
  FiRefreshCw,
  FiShuffle,
} from 'react-icons/fi'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { apiFetch } from '../../utils/apiFetch'

const DEFAULT_NAMESPACE = 'kubeevalhub-agent'
const DEFAULT_AGENT_IMAGE = 'docker.io/lewaldenko/kubeevalhubagent:1.0.0'
const DEFAULT_BACKEND_URL = 'http://host.minikube.internal:5000'

export default function DashboardPage() {
  const [clusters, setClusters] = useState([])
  const [agents, setAgents] = useState([])

  const [isLoadingClusters, setIsLoadingClusters] = useState(false)
  const [isLoadingAgents, setIsLoadingAgents] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const [clustersError, setClustersError] = useState('')
  const [agentsError, setAgentsError] = useState('')
  const [assignError, setAssignError] = useState('')

  const [showCreateClusterModal, setShowCreateClusterModal] = useState(false)
  const [clusterName, setClusterName] = useState('')
  const [isCreatingCluster, setIsCreatingCluster] = useState(false)
  const [createClusterError, setCreateClusterError] = useState('')

  const [showCreateAgentModal, setShowCreateAgentModal] = useState(false)
  const [agentName, setAgentName] = useState('')
  const [scrapeInterval, setScrapeInterval] = useState(30)
  const [isCreatingAgent, setIsCreatingAgent] = useState(false)
  const [createAgentError, setCreateAgentError] = useState('')
  const [createdAgentResult, setCreatedAgentResult] = useState(null)

  const [copiedToken, setCopiedToken] = useState(false)
  const [installCommandCopied, setInstallCommandCopied] = useState(false)
  const [manifestCopied, setManifestCopied] = useState(false)

  const [togglingAgentId, setTogglingAgentId] = useState('')
  const [deletingAgentId, setDeletingAgentId] = useState('')
  const [assigningClusterId, setAssigningClusterId] = useState('')
  const [selectedAgentByCluster, setSelectedAgentByCluster] = useState({})
  const [deletingClusterId, setDeletingClusterId] = useState('')

  const [clusterRolesByCluster, setClusterRolesByCluster] = useState({})
  const [roleEmailByCluster, setRoleEmailByCluster] = useState({})
  const [roleValueByCluster, setRoleValueByCluster] = useState({})
  const [roleErrorByCluster, setRoleErrorByCluster] = useState({})
  const [isLoadingRolesByCluster, setIsLoadingRolesByCluster] = useState({})
  const [isSavingRoleByCluster, setIsSavingRoleByCluster] = useState({})
  const [removingRoleKey, setRemovingRoleKey] = useState('')
  const [expandedAccessByCluster, setExpandedAccessByCluster] = useState({})

  const { isAuthenticated, isReady } = useAuth()
  const navigate = useNavigate()

  const getInstallConfig = () => {
    return {
      backendUrl: DEFAULT_BACKEND_URL,
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

  const buildInstallCommand = ({ token, scrapeIntervalSeconds }) => {
    const { namespace } = getInstallConfig()
    const manifest = buildInstallManifest({ token, scrapeIntervalSeconds })

    return `cat <<'EOF' | kubectl apply -f -
${manifest}
EOF

kubectl -n ${namespace} rollout status deployment/kubeevalhub-agent`
  }

  const loadClusters = async ({ silent = false } = {}) => {
    if (!isAuthenticated) return

    if (!silent) {
      setIsLoadingClusters(true)
    }

    setClustersError('')

    try {
      const response = await apiFetch('/api/clusters')
      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to load clusters.')
      }

      const nextClusters = Array.isArray(data) ? data : []
      setClusters(nextClusters)

      setSelectedAgentByCluster((prev) => {
        const next = { ...prev }

        nextClusters.forEach((cluster) => {
          if (cluster.agentId) {
            next[cluster.id] = cluster.agentId
          } else if (!(cluster.id in next)) {
            next[cluster.id] = ''
          }
        })

        return next
      })

      setRoleValueByCluster((prev) => {
        const next = { ...prev }
        nextClusters.forEach((cluster) => {
          if (!(cluster.id in next)) {
            next[cluster.id] = 'viewer'
          }
        })
        return next
      })
    } catch (error) {
      setClustersError(error.message || 'Failed to load clusters.')
    } finally {
      if (!silent) {
        setIsLoadingClusters(false)
      }
    }
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
    setAssignError('')

    try {
      await Promise.all([
        loadClusters({ silent: true }),
        loadAgents({ silent: true }),
      ])
    } finally {
      setIsRefreshing(false)
    }
  }

  useEffect(() => {
    if (!isReady || !isAuthenticated) return

    loadClusters()
    loadAgents()
  }, [isReady, isAuthenticated])

  const loadClusterRoles = async (clusterId) => {
    setIsLoadingRolesByCluster((prev) => ({ ...prev, [clusterId]: true }))
    setRoleErrorByCluster((prev) => ({ ...prev, [clusterId]: '' }))

    try {
      const response = await apiFetch(`/api/clusters/${clusterId}/roles`)
      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to load cluster access.')
      }

      setClusterRolesByCluster((prev) => ({
        ...prev,
        [clusterId]: Array.isArray(data) ? data : [],
      }))
    } catch (error) {
      setRoleErrorByCluster((prev) => ({
        ...prev,
        [clusterId]: error.message || 'Failed to load cluster access.',
      }))
    } finally {
      setIsLoadingRolesByCluster((prev) => ({ ...prev, [clusterId]: false }))
    }
  }

  const toggleAccessPanel = async (clusterId) => {
    const isExpanded = Boolean(expandedAccessByCluster[clusterId])

    if (isExpanded) {
      setExpandedAccessByCluster((prev) => ({ ...prev, [clusterId]: false }))
      return
    }

    setExpandedAccessByCluster((prev) => ({ ...prev, [clusterId]: true }))

    if (!clusterRolesByCluster[clusterId]) {
      await loadClusterRoles(clusterId)
    }
  }

  const handleGrantClusterAccess = async (clusterId) => {
    const email = String(roleEmailByCluster[clusterId] || '').trim().toLowerCase()
    const role = roleValueByCluster[clusterId] || 'viewer'

    if (!email) {
      setRoleErrorByCluster((prev) => ({
        ...prev,
        [clusterId]: 'Email is required.',
      }))
      return
    }

    setIsSavingRoleByCluster((prev) => ({ ...prev, [clusterId]: true }))
    setRoleErrorByCluster((prev) => ({ ...prev, [clusterId]: '' }))

    try {
      const response = await apiFetch(`/api/clusters/${clusterId}/roles`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
          role,
        }),
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to grant cluster access.')
      }

      setRoleEmailByCluster((prev) => ({ ...prev, [clusterId]: '' }))
      await loadClusterRoles(clusterId)
    } catch (error) {
      setRoleErrorByCluster((prev) => ({
        ...prev,
        [clusterId]: error.message || 'Failed to grant cluster access.',
      }))
    } finally {
      setIsSavingRoleByCluster((prev) => ({ ...prev, [clusterId]: false }))
    }
  }

  const handleRemoveClusterAccess = async (clusterId, userId) => {
    const key = `${clusterId}:${userId}`
    setRemovingRoleKey(key)
    setRoleErrorByCluster((prev) => ({ ...prev, [clusterId]: '' }))

    try {
      const response = await apiFetch(`/api/clusters/${clusterId}/roles/${userId}`, {
        method: 'DELETE',
      })

      let data = null
      if (response.status !== 204) {
        data = await response.json().catch(() => null)
      }

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to remove cluster access.')
      }

      await loadClusterRoles(clusterId)
    } catch (error) {
      setRoleErrorByCluster((prev) => ({
        ...prev,
        [clusterId]: error.message || 'Failed to remove cluster access.',
      }))
    } finally {
      setRemovingRoleKey('')
    }
  }

  const openCreateClusterModal = () => {
    setClusterName('')
    setCreateClusterError('')
    setShowCreateClusterModal(true)
  }

  const closeCreateClusterModal = () => {
    if (isCreatingCluster) return
    setShowCreateClusterModal(false)
    setClusterName('')
    setCreateClusterError('')
  }

  const handleGenerateRandomClusterName = () => {
    setClusterName(generateRandomClusterName())
  }

  const handleCreateCluster = async (e) => {
    e.preventDefault()

    const trimmedName = clusterName.trim()
    if (!trimmedName) {
      setCreateClusterError('Cluster name is required.')
      return
    }

    setIsCreatingCluster(true)
    setCreateClusterError('')

    try {
      const response = await apiFetch('/api/clusters', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          clusterName: trimmedName,
        }),
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to create cluster.')
      }

      closeCreateClusterModal()
      await loadClusters({ silent: true })
    } catch (error) {
      setCreateClusterError(error.message || 'Failed to create cluster.')
    } finally {
      setIsCreatingCluster(false)
    }
  }

  const openCreateAgentModal = () => {
    setAgentName('')
    setScrapeInterval(30)
    setCreateAgentError('')
    setCreatedAgentResult(null)
    setCopiedToken(false)
    setInstallCommandCopied(false)
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
    setCopiedToken(false)
    setInstallCommandCopied(false)
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
    setCopiedToken(false)
    setInstallCommandCopied(false)
    setManifestCopied(false)

    try {
      const response = await apiFetch('/api/agents', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
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

  const handleAssignAgent = async (clusterId) => {
    const agentId = selectedAgentByCluster[clusterId]

    if (!clusterId || !agentId) {
      setAssignError('Select an agent before assigning.')
      return
    }

    setAssigningClusterId(clusterId)
    setAssignError('')

    try {
      const response = await apiFetch(`/api/clusters/${clusterId}/assign-agent`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          agentId,
        }),
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to assign agent.')
      }

      await loadClusters({ silent: true })
    } catch (error) {
      setAssignError(error.message || 'Failed to assign agent.')
    } finally {
      setAssigningClusterId('')
    }
  }

  const handleDeleteCluster = async (clusterId, clusterNameValue) => {
    if (!isAuthenticated) return

    const confirmed = window.confirm(
      `Delete cluster "${clusterNameValue}"? This will remove the cluster record and its stored data.`
    )

    if (!confirmed) return

    setDeletingClusterId(clusterId)
    setClustersError('')
    setAssignError('')

    try {
      const response = await apiFetch(`/api/clusters/${clusterId}`, {
        method: 'DELETE',
      })

      let data = null
      if (response.status !== 204) {
        data = await response.json().catch(() => null)
      }

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to delete cluster.')
      }

      setClusters((prev) => prev.filter((cluster) => cluster.id !== clusterId))

      setSelectedAgentByCluster((prev) => {
        const next = { ...prev }
        delete next[clusterId]
        return next
      })

      setClusterRolesByCluster((prev) => {
        const next = { ...prev }
        delete next[clusterId]
        return next
      })
    } catch (error) {
      setClustersError(error.message || 'Failed to delete cluster.')
    } finally {
      setDeletingClusterId('')
    }
  }

  const handleDeleteAgent = async (agentId, agentNameValue) => {
    if (!isAuthenticated) return

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
        setCopiedToken(false)
        setInstallCommandCopied(false)
        setManifestCopied(false)
      }

      await Promise.all([
        loadAgents({ silent: true }),
        loadClusters({ silent: true }),
      ])
    } catch (error) {
      setAgentsError(error.message || 'Failed to delete agent.')
    } finally {
      setDeletingAgentId('')
    }
  }

  const handleToggleAgent = async (agentId, enabled) => {
    if (!isAuthenticated) return

    setTogglingAgentId(agentId)
    setAgentsError('')

    try {
      const response = await apiFetch(`/api/agents/${agentId}/enabled`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
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
      setCopiedToken(true)
      window.setTimeout(() => setCopiedToken(false), 2000)
    } catch {
      setCopiedToken(false)
    }
  }

  const handleCopyInstallCommand = async () => {
    const tokenValue = createdAgentResult?.token
    if (!tokenValue) return

    try {
      const command = buildInstallCommand({
        token: tokenValue,
        scrapeIntervalSeconds: scrapeInterval,
      })

      await navigator.clipboard.writeText(command)
      setInstallCommandCopied(true)
      window.setTimeout(() => setInstallCommandCopied(false), 2000)
    } catch (error) {
      setCreateAgentError(error.message || 'Failed to build install command.')
      setInstallCommandCopied(false)
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

  const handleShowMetrics = (clusterId) => {
    navigate(`/dashboard/clusters/${clusterId}/metrics`)
  }

  const handleShowInventory = (clusterId) => {
    navigate(`/dashboard/clusters/${clusterId}/inventory`)
  }

  const handleShowRecommendations = (clusterId) => {
    navigate(`/dashboard/clusters/${clusterId}/recommendations`)
  }

  const assignableAgents = agents.filter((agent) => agent.enabled)

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Dashboard</h1>
          <p>Create clusters, create agents, assign them, inspect cluster data, and manage access.</p>
        </div>

        <button
          type="button"
          className="secondary-button"
          onClick={refreshAll}
          disabled={isRefreshing || isLoadingClusters || isLoadingAgents}
        >
          <FiRefreshCw />
          <span>{isRefreshing ? 'Refreshing...' : 'Refresh'}</span>
        </button>
      </div>

      <section className="dashboard-cards">
        <div className="dashboard-card dashboard-card-primary agents-panel">
          <div className="agents-panel-header">
            <div>
              <h3>Clusters</h3>
              <p>Create cluster records before assigning agents.</p>
            </div>

            <button
              className="create-agent-button create-agent-icon-button"
              type="button"
              onClick={openCreateClusterModal}
              aria-label="Create cluster"
              title="Create cluster"
            >
              <FiPlus />
            </button>
          </div>
        </div>

        <div className="dashboard-card dashboard-card-primary agents-panel">
          <div className="agents-panel-header">
            <div>
              <h3>Agents</h3>
              <p>Create agents and retrieve their connection tokens.</p>
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
              <h3>Clusters</h3>
              <p>Assign agents, access data, and manage delegated access.</p>
            </div>
          </div>

          {(clustersError || assignError) && (
            <div className="form-error">{clustersError || assignError}</div>
          )}

          {isLoadingClusters ? (
            <p>Loading clusters...</p>
          ) : clusters.length === 0 ? (
            <p>No clusters created yet.</p>
          ) : (
            <div className="agents-list">
              {clusters.map((cluster) => {
                const isAssigned = Boolean(cluster.agentId)
                const selectedAgentId = selectedAgentByCluster[cluster.id] || ''
                const assignedAgentName = isAssigned
                  ? agents.find((agent) => agent.id === cluster.agentId)?.name || cluster.agentId
                  : 'Unassigned'

                const myRole = cluster.myRole || 'none'
                const canManageCluster = myRole === 'admin'
                const canUseAnalysis = myRole === 'admin' || myRole === 'operator'
                const canViewReadOnly = myRole === 'admin' || myRole === 'operator' || myRole === 'viewer'

                const clusterRoles = clusterRolesByCluster[cluster.id] || []
                const roleError = roleErrorByCluster[cluster.id] || ''
                const isLoadingRoles = Boolean(isLoadingRolesByCluster[cluster.id])
                const isSavingRole = Boolean(isSavingRoleByCluster[cluster.id])
                const isAccessExpanded = Boolean(expandedAccessByCluster[cluster.id])

                return (
                  <div key={cluster.id} className="agent-list-item">
                    <div className="agent-list-main">
                      <h4>{cluster.clusterName}</h4>

                      <div className="cluster-meta-list">
                        <div className="cluster-meta-row">
                          <span>ID</span>
                          <strong>{cluster.id}</strong>
                        </div>

                        <div className="cluster-meta-row">
                          <span>Cluster UID</span>
                          <strong>{cluster.clusterUid || 'Not discovered yet'}</strong>
                        </div>

                        <div className="cluster-meta-row">
                          <span>Kubernetes</span>
                          <strong>{cluster.kubeVersion || 'Not reported yet'}</strong>
                        </div>

                        <div className="cluster-meta-row">
                          <span>Distribution</span>
                          <strong>{cluster.distribution || 'Not reported yet'}</strong>
                        </div>

                        <div className="cluster-meta-row">
                          <span>Assigned agent</span>
                          <strong>{assignedAgentName}</strong>
                        </div>

                        <div className="cluster-meta-row">
                          <span>Last heartbeat</span>
                          <strong>
                            {cluster.lastHeartbeatAt
                              ? new Date(cluster.lastHeartbeatAt).toLocaleString()
                              : 'Never'}
                          </strong>
                        </div>

                        <div className="cluster-meta-row">
                          <span>My access</span>
                          <strong>{myRole}</strong>
                        </div>
                      </div>
                    </div>

                    <div className="cluster-side-panel">
                      <div className="cluster-side-header">
                        <span className={`agent-status-badge status-${cluster.status || 'offline'}`}>
                          {cluster.status || 'unknown'}
                        </span>

                        <div
                          className={
                            isAssigned
                              ? 'state-badge state-badge-enabled'
                              : 'state-badge state-badge-disabled'
                          }
                        >
                          {isAssigned ? 'Agent assigned' : 'No agent assigned'}
                        </div>
                      </div>

                      {canManageCluster && (
                        <div className="assignment-card">
                          <div className="assignment-card-header">
                            <div>
                              <h5>Agent assignment</h5>
                              <p>
                                {isAssigned
                                  ? 'Change the connected agent for this cluster.'
                                  : 'Select an available agent to connect this cluster.'}
                              </p>
                            </div>
                          </div>

                          <div className="assignment-current">
                            <span className="assignment-current-label">Current</span>
                            <span className="assignment-current-value">
                              {assignedAgentName}
                            </span>
                          </div>

                          <div className="assignment-controls">
                            <label className="assignment-select-group">
                              <span>Select agent</span>
                              <select
                                className="assignment-select"
                                value={selectedAgentId}
                                onChange={(e) =>
                                  setSelectedAgentByCluster((prev) => ({
                                    ...prev,
                                    [cluster.id]: e.target.value,
                                  }))
                                }
                                disabled={
                                  assigningClusterId === cluster.id || deletingClusterId === cluster.id
                                }
                              >
                                <option value="">Select agent</option>
                                {assignableAgents.map((agent) => (
                                  <option key={agent.id} value={agent.id}>
                                    {agent.name}
                                  </option>
                                ))}
                              </select>
                            </label>

                            <button
                              type="button"
                              className="primary-button assignment-button"
                              onClick={() => handleAssignAgent(cluster.id)}
                              disabled={
                                !selectedAgentId ||
                                assigningClusterId === cluster.id ||
                                deletingClusterId === cluster.id
                              }
                            >
                              {assigningClusterId === cluster.id
                                ? 'Assigning...'
                                : isAssigned
                                  ? 'Update assignment'
                                  : 'Assign agent'}
                            </button>
                          </div>
                        </div>
                      )}

                      {canManageCluster && (
                        <div className="assignment-card">
                          <div className="assignment-card-header">
                            <div>
                              <h5>Cluster access</h5>
                              <p>Grant another user viewer or operator access to this cluster.</p>
                            </div>
                          </div>

                          <div className="assignment-controls">
                            <label className="assignment-select-group">
                              <span>User email</span>
                              <input
                                type="email"
                                value={roleEmailByCluster[cluster.id] || ''}
                                onChange={(e) =>
                                  setRoleEmailByCluster((prev) => ({
                                    ...prev,
                                    [cluster.id]: e.target.value,
                                  }))
                                }
                                placeholder="Enter user email"
                                disabled={isSavingRole || deletingClusterId === cluster.id}
                              />
                            </label>

                            <label className="assignment-select-group">
                              <span>Role</span>
                              <select
                                className="assignment-select"
                                value={roleValueByCluster[cluster.id] || 'viewer'}
                                onChange={(e) =>
                                  setRoleValueByCluster((prev) => ({
                                    ...prev,
                                    [cluster.id]: e.target.value,
                                  }))
                                }
                                disabled={isSavingRole || deletingClusterId === cluster.id}
                              >
                                <option value="viewer">viewer</option>
                                <option value="operator">operator</option>
                              </select>
                            </label>

                            <button
                              type="button"
                              className="primary-button assignment-button"
                              onClick={() => handleGrantClusterAccess(cluster.id)}
                              disabled={isSavingRole || deletingClusterId === cluster.id}
                            >
                              {isSavingRole ? 'Saving...' : 'Grant access'}
                            </button>
                          </div>

                          <div className="cluster-actions">
                            <button
                              type="button"
                              className="secondary-button"
                              onClick={() => toggleAccessPanel(cluster.id)}
                              disabled={isSavingRole}
                            >
                              {isAccessExpanded ? 'Hide access list' : 'Show access list'}
                            </button>
                          </div>

                          {roleError && <div className="form-error">{roleError}</div>}

                          {isAccessExpanded && (
                            <div className="cluster-access-list">
                              {isLoadingRoles ? (
                                <p>Loading access...</p>
                              ) : clusterRoles.length === 0 ? (
                                <p>No delegated access configured.</p>
                              ) : (
                                <div className="agents-list">
                                  {clusterRoles.map((accessItem) => {
                                    const removeKey = `${cluster.id}:${accessItem.userId}`

                                    return (
                                      <div key={`${accessItem.clusterId}:${accessItem.userId}`} className="agent-list-item">
                                        <div className="agent-list-main">
                                          <div className="cluster-meta-list">
                                            <div className="cluster-meta-row">
                                              <span>User ID</span>
                                              <strong>{accessItem.userId}</strong>
                                            </div>
                                            <div className="cluster-meta-row">
                                              <span>Role</span>
                                              <strong>{accessItem.role}</strong>
                                            </div>
                                            <div className="cluster-meta-row">
                                              <span>Granted at</span>
                                              <strong>
                                                {accessItem.createdAt
                                                  ? new Date(accessItem.createdAt).toLocaleString()
                                                  : 'Unknown'}
                                              </strong>
                                            </div>
                                          </div>
                                        </div>

                                        <div className="agent-list-side">
                                          <button
                                            type="button"
                                            className="secondary-button"
                                            onClick={() => handleRemoveClusterAccess(cluster.id, accessItem.userId)}
                                            disabled={removingRoleKey === removeKey}
                                          >
                                            {removingRoleKey === removeKey ? 'Removing...' : 'Remove'}
                                          </button>
                                        </div>
                                      </div>
                                    )
                                  })}
                                </div>
                              )}
                            </div>
                          )}
                        </div>
                      )}

                      <div className="cluster-actions cluster-actions-grid">
                        <button
                          type="button"
                          className="dashboard-nav-button"
                          onClick={() => handleShowInventory(cluster.id)}
                          disabled={!isAssigned || !canViewReadOnly}
                        >
                          Show inventory
                        </button>

                        <button
                          type="button"
                          className="dashboard-nav-button"
                          onClick={() => handleShowRecommendations(cluster.id)}
                          disabled={!isAssigned || !canUseAnalysis}
                        >
                          Show recommendations
                        </button>

                        <button
                          type="button"
                          className="dashboard-nav-button is-primary"
                          onClick={() => handleShowMetrics(cluster.id)}
                          disabled={!isAssigned || !canViewReadOnly}
                        >
                          Show metrics
                        </button>

                        {canManageCluster && (
                          <button
                            type="button"
                            className="secondary-button"
                            onClick={() => handleDeleteCluster(cluster.id, cluster.clusterName)}
                            disabled={deletingClusterId === cluster.id || assigningClusterId === cluster.id}
                          >
                            {deletingClusterId === cluster.id ? 'Deleting...' : 'Delete cluster'}
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </section>

      <section className="dashboard-cards dashboard-cards-single">
        <div className="dashboard-card agents-panel">
          <div className="agents-panel-header">
            <div>
              <h3>Existing agents</h3>
              <p>View current state and enable or disable ingestion.</p>
            </div>
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

      {showCreateClusterModal && (
        <div className="modal-backdrop" onClick={closeCreateClusterModal}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h3>Create cluster</h3>
                <p>Create a cluster record before assigning an agent.</p>
              </div>

              <button className="icon-button" type="button" onClick={closeCreateClusterModal}>
                <FiX />
              </button>
            </div>

            <form onSubmit={handleCreateCluster} className="agent-form">
              <label className="field-group">
                <span>Cluster name</span>
                <div className="token-row">
                  <input
                    type="text"
                    value={clusterName}
                    onChange={(e) => setClusterName(e.target.value)}
                    placeholder="Enter cluster name"
                    maxLength={128}
                    disabled={isCreatingCluster}
                  />
                  <button
                    type="button"
                    className="copy-button"
                    onClick={handleGenerateRandomClusterName}
                    disabled={isCreatingCluster}
                  >
                    <FiShuffle />
                    <span>Random</span>
                  </button>
                </div>
              </label>

              {createClusterError && <div className="form-error">{createClusterError}</div>}

              <div className="modal-actions">
                <button
                  type="button"
                  className="secondary-button"
                  onClick={closeCreateClusterModal}
                >
                  Cancel
                </button>

                <button type="submit" className="primary-button" disabled={isCreatingCluster}>
                  {isCreatingCluster ? 'Creating...' : 'Create cluster'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

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
                  <p>Run the installer on the target cluster. The token is shown only once.</p>
                </div>

                <div className="result-meta">
                  <div>
                    <strong>Name:</strong> {createdAgentResult.agent?.name}
                  </div>
                  <div>
                    <strong>ID:</strong> {createdAgentResult.agent?.id}
                  </div>
                  <div>
                    <strong>Backend URL:</strong> {getInstallConfig().backendUrl}
                  </div>
                  <div>
                    <strong>Image:</strong> {getInstallConfig().image}
                  </div>
                </div>

                <div className="token-block">
                  <label>Agent token</label>
                  <div className="token-row">
                    <code>{createdAgentResult.token}</code>
                    <button type="button" className="copy-button" onClick={handleCopyToken}>
                      {copiedToken ? <FiCheck /> : <FiCopy />}
                      <span>{copiedToken ? 'Copied' : 'Copy'}</span>
                    </button>
                  </div>
                </div>

                {createAgentError && <div className="form-error">{createAgentError}</div>}

                <div className="copy-actions-grid">
                  <button
                    type="button"
                    className="primary-button"
                    onClick={handleCopyInstallCommand}
                  >
                    {installCommandCopied ? <FiCheck /> : <FiCopy />}
                    <span>
                      {installCommandCopied ? 'Copied install command' : 'Copy install command'}
                    </span>
                  </button>

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

function generateRandomClusterName() {
  const adjectives = [
    'silent',
    'rapid',
    'steady',
    'bright',
    'north',
    'east',
    'prime',
    'crystal',
    'ember',
    'iron',
  ]

  const nouns = [
    'falcon',
    'harbor',
    'summit',
    'forge',
    'meadow',
    'anchor',
    'matrix',
    'horizon',
    'ridge',
    'vertex',
  ]

  const adjective = adjectives[Math.floor(Math.random() * adjectives.length)]
  const noun = nouns[Math.floor(Math.random() * nouns.length)]
  const suffix = Math.floor(100 + Math.random() * 900)

  return `${adjective}-${noun}-${suffix}`
}