import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { apiFetch } from '../../utils/apiFetch'

export default function ClustersPage() {
  const [clusters, setClusters] = useState([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const { isAuthenticated, isReady } = useAuth()
  const navigate = useNavigate()

  const loadClusters = async () => {
    if (!isAuthenticated) return

    setIsLoading(true)
    setError('')

    try {
      const res = await apiFetch('/api/clusters')
      const data = await res.json().catch(() => null)

      if (!res.ok) {
        throw new Error(data?.error || 'Failed to load clusters')
      }

      setClusters(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message || 'Failed to load clusters')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (!isReady || !isAuthenticated) return

    loadClusters()

    const id = setInterval(loadClusters, 15000)
    return () => clearInterval(id)
  }, [isReady, isAuthenticated])

  const handleShowMetrics = (agentId) => {
    navigate(`/dashboard/clusters/${agentId}/metrics`)
  }

  const handleShowInventory = (agentId) => {
    navigate(`/dashboard/clusters/${agentId}/inventory`)
  }

  const handleShowRecommendations = (agentId) => {
    navigate(`/dashboard/clusters/${agentId}/recommendations`)
  }

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Clusters</h1>
          <p>Inspect connected Kubernetes clusters and their status.</p>
        </div>
      </div>

      <section className="dashboard-cards dashboard-cards-single">
        <div className="dashboard-card">
          {error && <div className="form-error">{error}</div>}

          {isLoading ? (
            <p>Loading clusters...</p>
          ) : clusters.length === 0 ? (
            <p>No clusters connected yet.</p>
          ) : (
            <div className="agents-list">
              {clusters.map((c) => (
                <div key={c.clusterUid} className="agent-list-item">
                  <div className="agent-list-main">
                    <h4>{c.clusterName}</h4>
                    <p>UID: {c.clusterUid}</p>
                    <p>Kubernetes: {c.kubeVersion}</p>
                    <p>Distribution: {c.distribution}</p>
                    <p>
                      Last heartbeat:{' '}
                      {c.lastHeartbeatAt
                        ? new Date(c.lastHeartbeatAt).toLocaleString()
                        : 'Never'}
                    </p>
                  </div>

                  <div className="agent-list-side">
                    <span className={`agent-status-badge status-${c.status}`}>
                      {c.status}
                    </span>

                    <div className="cluster-actions">
                      <button
                        type="button"
                        className="dashboard-nav-button"
                        onClick={() => handleShowInventory(c.agentId)}
                      >
                        Show inventory
                      </button>

                      <button
                        type="button"
                        className="dashboard-nav-button"
                        onClick={() => handleShowRecommendations(c.agentId)}
                      >
                        Show recommendations
                      </button>

                      <button
                        type="button"
                        className="dashboard-nav-button is-primary"
                        onClick={() => handleShowMetrics(c.agentId)}
                      >
                        Show metrics
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </section>
    </>
  )
}