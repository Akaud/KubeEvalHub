import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { apiFetch } from '../../utils/apiFetch'

function formatFullDateTime(value) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function formatCPU(millicores) {
  if (millicores == null || !Number.isFinite(millicores)) return '—'
  return `${millicores}m`
}

function formatBytes(bytes) {
  if (bytes == null || !Number.isFinite(bytes)) return '—'
  if (bytes === 0) return '0 B'

  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB']
  let value = bytes
  let unitIndex = 0

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }

  return `${value.toFixed(value >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function flattenWorkloads(inventoryPayload) {
  if (!inventoryPayload) return []

  const deployments = (inventoryPayload.deployments || []).map((item) => ({
    kind: 'Deployment',
    ...item,
  }))

  const statefulSets = (inventoryPayload.statefulSets || []).map((item) => ({
    kind: 'StatefulSet',
    ...item,
  }))

  const daemonSets = (inventoryPayload.daemonSets || []).map((item) => ({
    kind: 'DaemonSet',
    ...item,
  }))

  return [...deployments, ...statefulSets, ...daemonSets]
}

function flattenPodContainers(inventoryPayload) {
  if (!inventoryPayload?.pods?.length) return []

  return inventoryPayload.pods.flatMap((pod) =>
    (pod.containers || []).map((container) => ({
      namespace: pod.namespace,
      podName: pod.name,
      nodeName: pod.nodeName,
      phase: pod.phase,
      ownerKind: pod.controllerKind,
      ownerName: pod.controllerName,
      containerName: container.name,
      image: container.image,
      cpuRequestMillicores: container.cpuRequestMillicores,
      cpuLimitMillicores: container.cpuLimitMillicores,
      memoryRequestBytes: container.memoryRequestBytes,
      memoryLimitBytes: container.memoryLimitBytes,
    }))
  )
}

function getPhaseBadgeClass(phase) {
  const normalized = String(phase || '').toLowerCase()
  if (normalized === 'running') return 'status-pill status-pill-success'
  if (normalized === 'pending') return 'status-pill status-pill-warning'
  if (normalized === 'failed') return 'status-pill status-pill-danger'
  return 'status-pill'
}

export default function ClusterInventoryDetailPage() {
  const { clusterId } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated, isReady } = useAuth()

  const [cluster, setCluster] = useState(null)
  const [inventory, setInventory] = useState(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [error, setError] = useState('')

  const loadClusterAndInventory = useCallback(async (refresh = false) => {
    if (!isAuthenticated || !clusterId) return

    if (refresh) {
      setIsRefreshing(true)
    } else {
      setIsLoading(true)
    }

    setError('')

    try {
      const clustersRes = await apiFetch('/api/clusters')
      const clustersData = await clustersRes.json().catch(() => null)

      if (!clustersRes.ok) {
        throw new Error(clustersData?.error || 'Failed to load clusters')
      }

      const clusters = Array.isArray(clustersData) ? clustersData : []
      const currentCluster = clusters.find((item) => item.id === clusterId) || null
      setCluster(currentCluster)

      const inventoryRes = await apiFetch(`/api/clusters/${clusterId}/inventory/latest`)

      if (inventoryRes.status === 404) {
        setInventory(null)
        setError('')
        return
      }

      const inventoryData = await inventoryRes.json().catch(() => null)

      if (!inventoryRes.ok) {
        throw new Error(inventoryData?.error || 'Failed to load inventory')
      }

      setInventory(inventoryData)
    } catch (err) {
      setInventory(null)
      setError(err.message || 'Failed to load inventory')
    } finally {
      setIsLoading(false)
      setIsRefreshing(false)
    }
  }, [isAuthenticated, clusterId])

  useEffect(() => {
    if (!isReady || !isAuthenticated) return
    loadClusterAndInventory(false)
  }, [isReady, isAuthenticated, loadClusterAndInventory])

  const inventoryPayload = inventory?.inventory || null

  const summary = useMemo(() => {
    if (!inventoryPayload) {
      return {
        namespaceCount: 0,
        nodeCount: 0,
        deploymentCount: 0,
        statefulSetCount: 0,
        daemonSetCount: 0,
        podCount: 0,
        containerCount: 0,
      }
    }

    const containerCount = (inventoryPayload.pods || []).reduce(
      (acc, pod) => acc + (pod.containers?.length || 0),
      0
    )

    return {
      namespaceCount: inventoryPayload.namespaces?.length || 0,
      nodeCount: inventoryPayload.nodes?.length || 0,
      deploymentCount: inventoryPayload.deployments?.length || 0,
      statefulSetCount: inventoryPayload.statefulSets?.length || 0,
      daemonSetCount: inventoryPayload.daemonSets?.length || 0,
      podCount: inventoryPayload.pods?.length || 0,
      containerCount,
    }
  }, [inventoryPayload])

  const workloads = useMemo(() => flattenWorkloads(inventoryPayload), [inventoryPayload])
  const podContainers = useMemo(() => flattenPodContainers(inventoryPayload), [inventoryPayload])

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>{cluster?.clusterName || 'Cluster inventory'}</h1>
          <p>Latest workload inventory and resource specification snapshot.</p>
        </div>
      </div>

      <section className="dashboard-cards dashboard-cards-single">
        <div className="dashboard-card inventory-page-card">
          <div className="metrics-page-toolbar">
            <div className="metrics-toolbar-left">
              <button
                type="button"
                className="dashboard-nav-button"
                onClick={() => navigate('/dashboard/clusters')}
              >
                Back to clusters
              </button>

              <button
                type="button"
                className="dashboard-nav-button is-primary"
                onClick={() => loadClusterAndInventory(true)}
                disabled={isLoading || isRefreshing}
              >
                {isRefreshing ? 'Refreshing...' : 'Refresh inventory'}
              </button>
            </div>
          </div>

          {error && <div className="form-error">{error}</div>}

          {isLoading ? (
            <p>Loading inventory...</p>
          ) : !inventoryPayload ? (
            <p>No inventory available.</p>
          ) : (
            <div className="inventory-sections">
              <section className="metrics-chart-card inventory-section">
                <div className="inventory-section-header">
                  <div>
                    <h3>Snapshot overview</h3>
                    <p>Collected at {formatFullDateTime(inventory?.collectedAt)}</p>
                  </div>
                </div>

                <div className="inventory-summary-grid">
                  <div className="inventory-summary-card">
                    <span className="inventory-summary-label">Namespaces</span>
                    <strong className="inventory-summary-value">{summary.namespaceCount}</strong>
                  </div>
                  <div className="inventory-summary-card">
                    <span className="inventory-summary-label">Nodes</span>
                    <strong className="inventory-summary-value">{summary.nodeCount}</strong>
                  </div>
                  <div className="inventory-summary-card">
                    <span className="inventory-summary-label">Deployments</span>
                    <strong className="inventory-summary-value">{summary.deploymentCount}</strong>
                  </div>
                  <div className="inventory-summary-card">
                    <span className="inventory-summary-label">StatefulSets</span>
                    <strong className="inventory-summary-value">{summary.statefulSetCount}</strong>
                  </div>
                  <div className="inventory-summary-card">
                    <span className="inventory-summary-label">DaemonSets</span>
                    <strong className="inventory-summary-value">{summary.daemonSetCount}</strong>
                  </div>
                  <div className="inventory-summary-card">
                    <span className="inventory-summary-label">Pods</span>
                    <strong className="inventory-summary-value">{summary.podCount}</strong>
                  </div>
                  <div className="inventory-summary-card">
                    <span className="inventory-summary-label">Containers</span>
                    <strong className="inventory-summary-value">{summary.containerCount}</strong>
                  </div>
                </div>
              </section>

              <section className="metrics-chart-card inventory-section">
                <div className="inventory-section-header">
                  <div>
                    <h3>Nodes</h3>
                    <p>Runtime and operating system details for each node.</p>
                  </div>
                </div>

                {!inventoryPayload.nodes?.length ? (
                  <p>No nodes available.</p>
                ) : (
                  <div className="table-responsive inventory-table-wrap">
                    <table className="dashboard-table inventory-table">
                      <thead>
                        <tr>
                          <th>Name</th>
                          <th>Kubelet</th>
                          <th>Runtime</th>
                          <th>OS</th>
                          <th>Architecture</th>
                        </tr>
                      </thead>
                      <tbody>
                        {inventoryPayload.nodes.map((node) => (
                          <tr key={node.uid || node.name}>
                            <td className="cell-code">{node.name || '—'}</td>
                            <td>{node.kubeletVersion || '—'}</td>
                            <td className="cell-code">{node.containerRuntimeVersion || '—'}</td>
                            <td>{node.osImage || node.operatingSystem || '—'}</td>
                            <td>{node.architecture || '—'}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </section>

              <section className="metrics-chart-card inventory-section">
                <div className="inventory-section-header">
                  <div>
                    <h3>Workloads</h3>
                    <p>Deployments, StatefulSets, and DaemonSets.</p>
                  </div>
                </div>

                {!workloads.length ? (
                  <p>No workloads available.</p>
                ) : (
                  <div className="table-responsive inventory-table-wrap">
                    <table className="dashboard-table inventory-table">
                      <thead>
                        <tr>
                          <th>Kind</th>
                          <th>Namespace</th>
                          <th>Name</th>
                          <th>Replicas</th>
                          <th>Containers</th>
                        </tr>
                      </thead>
                      <tbody>
                        {workloads.map((item) => (
                          <tr key={`${item.kind}-${item.uid || item.namespace}-${item.name}`}>
                            <td>{item.kind}</td>
                            <td>{item.namespace || '—'}</td>
                            <td className="cell-code">{item.name || '—'}</td>
                            <td>{item.replicas ?? '—'}</td>
                            <td>{item.containers?.length || 0}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </section>

              <section className="metrics-chart-card inventory-section">
                <div className="inventory-section-header">
                  <div>
                    <h3>Pods and containers</h3>
                    <p>Container resources, ownership, and placement.</p>
                  </div>
                </div>

                {!podContainers.length ? (
                  <p>No pod container inventory available.</p>
                ) : (
                  <div className="table-responsive inventory-table-wrap inventory-table-wrap-large">
                    <table className="dashboard-table inventory-table inventory-table-dense">
                      <thead>
                        <tr>
                          <th>Namespace</th>
                          <th>Pod</th>
                          <th>Container</th>
                          <th>Phase</th>
                          <th>Node</th>
                          <th>Owner</th>
                          <th>CPU Req</th>
                          <th>CPU Lim</th>
                          <th>Mem Req</th>
                          <th>Mem Lim</th>
                          <th>Image</th>
                        </tr>
                      </thead>
                      <tbody>
                        {podContainers.map((item, index) => (
                          <tr key={`${item.namespace}-${item.podName}-${item.containerName}-${index}`}>
                            <td>{item.namespace || '—'}</td>
                            <td className="cell-code">{item.podName || '—'}</td>
                            <td className="cell-code">{item.containerName || '—'}</td>
                            <td>
                              <span className={getPhaseBadgeClass(item.phase)}>
                                {item.phase || '—'}
                              </span>
                            </td>
                            <td className="cell-code">{item.nodeName || '—'}</td>
                            <td className="cell-code">
                              {item.ownerKind || item.ownerName
                                ? `${item.ownerKind || 'Unknown'} / ${item.ownerName || 'Unknown'}`
                                : '—'}
                            </td>
                            <td>{formatCPU(item.cpuRequestMillicores)}</td>
                            <td>{formatCPU(item.cpuLimitMillicores)}</td>
                            <td>{formatBytes(item.memoryRequestBytes)}</td>
                            <td>{formatBytes(item.memoryLimitBytes)}</td>
                            <td className="cell-code cell-truncate" title={item.image || '—'}>
                              {item.image || '—'}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </section>
            </div>
          )}
        </div>
      </section>
    </>
  )
}