import { useEffect, useRef, useState } from 'react'
import {
  FiPlus,
  FiRefreshCw,
  FiShuffle,
  FiX,
  FiMoreVertical,
} from 'react-icons/fi'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { apiFetch } from '../../utils/apiFetch'

export default function ClustersPage() {
  const [clusters, setClusters] = useState([])
  const [agents, setAgents] = useState([])
  const MAX_CLUSTERS = 3
  const clusterCount = clusters.length
  const canCreateCluster = clusterCount < MAX_CLUSTERS

  const [isLoadingClusters, setIsLoadingClusters] = useState(false)
  const [isLoadingAgents, setIsLoadingAgents] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const [clustersError, setClustersError] = useState('')
  const [assignError, setAssignError] = useState('')

  const [showCreateClusterModal, setShowCreateClusterModal] = useState(false)
  const [clusterName, setClusterName] = useState('')
  const [isCreatingCluster, setIsCreatingCluster] = useState(false)
  const [createClusterError, setCreateClusterError] = useState('')

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

  const [menuOpenClusterId, setMenuOpenClusterId] = useState('')
  const [activeAssignClusterId, setActiveAssignClusterId] = useState('')
  const [activeAccessClusterId, setActiveAccessClusterId] = useState('')

  const menuRef = useRef(null)

  const { isAuthenticated, isReady } = useAuth()
  const navigate = useNavigate()

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

    try {
      const response = await apiFetch('/api/agents')
      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to load agents.')
      }

      setAgents(Array.isArray(data) ? data : [])
    } catch (error) {
      setClustersError(error.message || 'Failed to load agents.')
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

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (!menuRef.current) return
      if (!menuRef.current.contains(event.target)) {
        setMenuOpenClusterId('')
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

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

  const openAssignAgentModal = (clusterId) => {
    setMenuOpenClusterId('')
    setAssignError('')
    setActiveAssignClusterId(clusterId)
  }

  const closeAssignAgentModal = () => {
    if (assigningClusterId) return
    setAssignError('')
    setActiveAssignClusterId('')
  }

  const openClusterAccessModal = async (clusterId) => {
    setMenuOpenClusterId('')
    setRoleErrorByCluster((prev) => ({ ...prev, [clusterId]: '' }))
    setActiveAccessClusterId(clusterId)

    if (!clusterRolesByCluster[clusterId]) {
      await loadClusterRoles(clusterId)
    }
  }

  const closeClusterAccessModal = () => {
    if (isSavingRoleByCluster[activeAccessClusterId]) return
    setActiveAccessClusterId('')
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
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, role }),
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
    if (!canCreateCluster) return

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

    if (!canCreateCluster) {
      setCreateClusterError(`You can create up to ${MAX_CLUSTERS} clusters.`)
      return
    }

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
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ clusterName: trimmedName }),
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
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agentId }),
      })

      const data = await response.json().catch(() => null)

      if (!response.ok) {
        throw new Error(data?.error || 'Failed to assign agent.')
      }

      await loadClusters({ silent: true })
      closeAssignAgentModal()
    } catch (error) {
      setAssignError(error.message || 'Failed to assign agent.')
    } finally {
      setAssigningClusterId('')
    }
  }

  const handleDeleteCluster = async (clusterId, clusterNameValue) => {
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

      if (activeAssignClusterId === clusterId) {
        setActiveAssignClusterId('')
      }

      if (activeAccessClusterId === clusterId) {
        setActiveAccessClusterId('')
      }
    } catch (error) {
      setClustersError(error.message || 'Failed to delete cluster.')
    } finally {
      setDeletingClusterId('')
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

  const activeAssignCluster = clusters.find((cluster) => cluster.id === activeAssignClusterId) || null
  const activeAccessCluster = clusters.find((cluster) => cluster.id === activeAccessClusterId) || null

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Clusters</h1>
          <p>Create cluster records, assign agents, manage delegated access, and open cluster views.</p>
          <p className="agents-counter">
            {clusterCount}/{MAX_CLUSTERS} clusters
          </p>
        </div>

        <div className="dashboard-header-actions">
          <button
            type="button"
            className="secondary-button"
            onClick={refreshAll}
            disabled={isRefreshing || isLoadingClusters || isLoadingAgents}
          >
            <FiRefreshCw />
            <span>{isRefreshing ? 'Refreshing...' : 'Refresh'}</span>
          </button>

          <button
            className="primary-button"
            type="button"
            onClick={openCreateClusterModal}
            disabled={!canCreateCluster}
            title={!canCreateCluster ? `Maximum of ${MAX_CLUSTERS} clusters reached` : undefined}
          >
            <FiPlus />
            <span>{canCreateCluster ? 'Create cluster' : 'Limit reached'}</span>
          </button>
        </div>
      </div>

      <section className="dashboard-cards dashboard-cards-single">
        <div className="dashboard-card agents-panel">
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
                const assignedAgentName = isAssigned
                  ? agents.find((agent) => agent.id === cluster.agentId)?.name || cluster.agentId
                  : 'Unassigned'

                const myRole = cluster.myRole || 'none'
                const canManageCluster = myRole === 'admin'
                const canUseAnalysis = myRole === 'admin' || myRole === 'operator'
                const canViewReadOnly =
                  myRole === 'admin' || myRole === 'operator' || myRole === 'viewer'

                const isMenuOpen = menuOpenClusterId === cluster.id

                return (
                  <div key={cluster.id} className="agent-list-item cluster-card-item">
                    <div className="cluster-card-menu" ref={isMenuOpen ? menuRef : null}>
                      {canManageCluster && (
                        <>
                          <button
                            type="button"
                            className="icon-button cluster-menu-trigger"
                            onClick={() =>
                              setMenuOpenClusterId((prev) => (prev === cluster.id ? '' : cluster.id))
                            }
                            aria-label="Open cluster actions"
                          >
                            <FiMoreVertical />
                          </button>

                          {isMenuOpen && (
                            <div className="cluster-menu-dropdown">
                              <button
                                type="button"
                                className="cluster-menu-item"
                                onClick={() => openAssignAgentModal(cluster.id)}
                              >
                                Assign agent
                              </button>

                              <button
                                type="button"
                                className="cluster-menu-item"
                                onClick={() => openClusterAccessModal(cluster.id)}
                              >
                                Cluster access
                              </button>
                            </div>
                          )}
                        </>
                      )}
                    </div>

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

                      <div className="cluster-actions cluster-actions-grid">
                        <button
                          type="button"
                          className="dashboard-nav-button"
                          onClick={() => handleShowInventory(cluster.id)}
                          disabled={!isAssigned || !canViewReadOnly || cluster.status !== 'online'}
                        >
                          Show inventory
                        </button>

                        <button
                          type="button"
                          className="dashboard-nav-button"
                          onClick={() => handleShowRecommendations(cluster.id)}
                          disabled={!isAssigned || !canViewReadOnly || cluster.status !== 'online'}
                        >
                          Show recommendations
                        </button>

                        <button
                          type="button"
                          className="dashboard-nav-button is-primary"
                          onClick={() => handleShowMetrics(cluster.id)}
                          disabled={!isAssigned || !canViewReadOnly || cluster.status !== 'online'}
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

      {activeAssignCluster && (
        <div className="modal-backdrop" onClick={closeAssignAgentModal}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h3>Assign agent</h3>
                <p>Connect an agent to {activeAssignCluster.clusterName}.</p>
              </div>

              <button className="icon-button" type="button" onClick={closeAssignAgentModal}>
                <FiX />
              </button>
            </div>

            {assignError && <div className="form-error">{assignError}</div>}

            <div className="assignment-card">
              <div className="assignment-current">
                <span className="assignment-current-label">Current</span>
                <span className="assignment-current-value">
                  {activeAssignCluster.agentId
                    ? agents.find((agent) => agent.id === activeAssignCluster.agentId)?.name ||
                      activeAssignCluster.agentId
                    : 'Unassigned'}
                </span>
              </div>

              <div className="assignment-controls">
                <label className="assignment-select-group">
                  <span>Select agent</span>
                  <select
                    className="assignment-select"
                    value={selectedAgentByCluster[activeAssignCluster.id] || ''}
                    onChange={(e) =>
                      setSelectedAgentByCluster((prev) => ({
                        ...prev,
                        [activeAssignCluster.id]: e.target.value,
                      }))
                    }
                    disabled={
                      assigningClusterId === activeAssignCluster.id ||
                      deletingClusterId === activeAssignCluster.id
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
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  className="secondary-button"
                  onClick={closeAssignAgentModal}
                  disabled={assigningClusterId === activeAssignCluster.id}
                >
                  Cancel
                </button>

                <button
                  type="button"
                  className="primary-button"
                  onClick={() => handleAssignAgent(activeAssignCluster.id)}
                  disabled={
                    !selectedAgentByCluster[activeAssignCluster.id] ||
                    assigningClusterId === activeAssignCluster.id ||
                    deletingClusterId === activeAssignCluster.id
                  }
                >
                  {assigningClusterId === activeAssignCluster.id
                    ? 'Assigning...'
                    : activeAssignCluster.agentId
                      ? 'Update assignment'
                      : 'Assign agent'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {activeAccessCluster && (
        <div className="modal-backdrop" onClick={closeClusterAccessModal}>
          <div className="modal-card modal-card-wide" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h3>Cluster access</h3>
                <p>Grant viewer or operator access to {activeAccessCluster.clusterName}.</p>
              </div>

              <button className="icon-button" type="button" onClick={closeClusterAccessModal}>
                <FiX />
              </button>
            </div>

            <div className="assignment-card">
              <div className="assignment-controls">
                <label className="assignment-select-group">
                  <span>User email</span>
                  <input
                    type="email"
                    value={roleEmailByCluster[activeAccessCluster.id] || ''}
                    onChange={(e) =>
                      setRoleEmailByCluster((prev) => ({
                        ...prev,
                        [activeAccessCluster.id]: e.target.value,
                      }))
                    }
                    placeholder="Enter user email"
                    disabled={
                      isSavingRoleByCluster[activeAccessCluster.id] ||
                      deletingClusterId === activeAccessCluster.id
                    }
                  />
                </label>

                <label className="assignment-select-group">
                  <span>Role</span>
                  <select
                    className="assignment-select"
                    value={roleValueByCluster[activeAccessCluster.id] || 'viewer'}
                    onChange={(e) =>
                      setRoleValueByCluster((prev) => ({
                        ...prev,
                        [activeAccessCluster.id]: e.target.value,
                      }))
                    }
                    disabled={
                      isSavingRoleByCluster[activeAccessCluster.id] ||
                      deletingClusterId === activeAccessCluster.id
                    }
                  >
                    <option value="viewer">viewer</option>
                    <option value="operator">operator</option>
                  </select>
                </label>
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  className="secondary-button"
                  onClick={closeClusterAccessModal}
                  disabled={isSavingRoleByCluster[activeAccessCluster.id]}
                >
                  Close
                </button>

                <button
                  type="button"
                  className="primary-button"
                  onClick={() => handleGrantClusterAccess(activeAccessCluster.id)}
                  disabled={
                    isSavingRoleByCluster[activeAccessCluster.id] ||
                    deletingClusterId === activeAccessCluster.id
                  }
                >
                  {isSavingRoleByCluster[activeAccessCluster.id] ? 'Saving...' : 'Grant access'}
                </button>
              </div>

              {roleErrorByCluster[activeAccessCluster.id] && (
                <div className="form-error">{roleErrorByCluster[activeAccessCluster.id]}</div>
              )}

              <div className="cluster-access-list">
                {isLoadingRolesByCluster[activeAccessCluster.id] ? (
                  <p>Loading access...</p>
                ) : (clusterRolesByCluster[activeAccessCluster.id] || []).length === 0 ? (
                  <p>No delegated access configured.</p>
                ) : (
                  <div className="agents-list">
                    {clusterRolesByCluster[activeAccessCluster.id].map((accessItem) => {
                      const removeKey = `${activeAccessCluster.id}:${accessItem.userId}`

                      return (
                        <div
                          key={`${accessItem.clusterId}:${accessItem.userId}`}
                          className="agent-list-item"
                        >
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
                              onClick={() =>
                                handleRemoveClusterAccess(activeAccessCluster.id, accessItem.userId)
                              }
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
            </div>
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