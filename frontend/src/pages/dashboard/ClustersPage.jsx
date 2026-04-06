import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '../../context/AuthContext'

export default function ClustersPage() {
  const [clusters, setClusters] = useState([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const { token } = useAuth()

  const authToken = useMemo(() => {
    return token || localStorage.getItem('token') || ''
  }, [token])

  const loadClusters = async () => {
    if (!authToken) return

    setIsLoading(true)
    setError('')

    try {
      const res = await fetch('/api/clusters', {
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      })

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
    loadClusters()

    const id = setInterval(loadClusters, 15000)
    return () => clearInterval(id)
  }, [authToken])

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