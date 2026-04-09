import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { apiFetch } from '../../utils/apiFetch'
import { useAuth } from '../../context/AuthContext'

function formatDateTimeInput(date) {
  const pad = (n) => String(n).padStart(2, '0')

  const year = date.getFullYear()
  const month = pad(date.getMonth() + 1)
  const day = pad(date.getDate())
  const hours = pad(date.getHours())
  const minutes = pad(date.getMinutes())

  return `${year}-${month}-${day}T${hours}:${minutes}`
}

function toRFC3339Local(value) {
  if (!value) return ''
  return new Date(value).toISOString()
}

function formatNumber(value, digits = 2) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return '—'
  }
  return Number(value).toFixed(digits)
}

function formatBytes(bytes) {
  if (bytes === null || bytes === undefined || Number.isNaN(Number(bytes))) {
    return '—'
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = Number(bytes)
  let unitIndex = 0

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }

  return `${value.toFixed(value >= 10 ? 0 : 2)} ${units[unitIndex]}`
}

function buildQuery(params) {
  const search = new URLSearchParams()

  Object.entries(params).forEach(([key, value]) => {
    if (value !== '' && value !== null && value !== undefined) {
      search.set(key, String(value))
    }
  })

  return search.toString()
}

export default function ClusterRecommendationsPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated, isReady } = useAuth()

  const defaultTo = useMemo(() => new Date(), [])
  const defaultFrom = useMemo(() => {
    const d = new Date()
    d.setDate(d.getDate() - 7)
    return d
  }, [])

  const [from, setFrom] = useState(formatDateTimeInput(defaultFrom))
  const [to, setTo] = useState(formatDateTimeInput(defaultTo))

  const [recommendations, setRecommendations] = useState(null)
  const [overProvisioned, setOverProvisioned] = useState(null)
  const [underProvisioned, setUnderProvisioned] = useState(null)
  const [capacity, setCapacity] = useState(null)

  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const loadData = async () => {
    if (!isAuthenticated || !id) return

    setIsLoading(true)
    setError('')

    try {
      const query = buildQuery({
        from: toRFC3339Local(from),
        to: toRFC3339Local(to),
      })

      const [
        recommendationsRes,
        overProvisionedRes,
        underProvisionedRes,
        capacityRes,
      ] = await Promise.all([
        apiFetch(`/api/clusters/${id}/recommendations?${query}`),
        apiFetch(`/api/clusters/${id}/analysis/overprovisioned?${query}`),
        apiFetch(`/api/clusters/${id}/analysis/underprovisioned?${query}`),
        apiFetch(`/api/clusters/${id}/capacity?${query}`),
      ])

      const [
        recommendationsData,
        overProvisionedData,
        underProvisionedData,
        capacityData,
      ] = await Promise.all([
        recommendationsRes.json().catch(() => null),
        overProvisionedRes.json().catch(() => null),
        underProvisionedRes.json().catch(() => null),
        capacityRes.json().catch(() => null),
      ])

      if (!recommendationsRes.ok) {
        throw new Error(recommendationsData?.error || 'Failed to load recommendations')
      }
      if (!overProvisionedRes.ok) {
        throw new Error(overProvisionedData?.error || 'Failed to load over-provisioned workloads')
      }
      if (!underProvisionedRes.ok) {
        throw new Error(underProvisionedData?.error || 'Failed to load under-provisioned workloads')
      }
      if (!capacityRes.ok) {
        throw new Error(capacityData?.error || 'Failed to load capacity')
      }

      setRecommendations(recommendationsData)
      setOverProvisioned(overProvisionedData)
      setUnderProvisioned(underProvisionedData)
      setCapacity(capacityData)
    } catch (err) {
      setError(err.message || 'Failed to load recommendations data')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (!isReady || !isAuthenticated) return

    loadData()
    const intervalId = setInterval(loadData, 30000)

    return () => clearInterval(intervalId)
  }, [isReady, isAuthenticated, id, from, to])

  const recommendationItems =
    recommendations?.items ||
    recommendations?.recommendations ||
    recommendations?.workloads ||
    []

  const overProvisionedItems =
    overProvisioned?.items ||
    overProvisioned?.workloads ||
    overProvisioned ||
    []

  const underProvisionedItems =
    underProvisioned?.items ||
    underProvisioned?.workloads ||
    underProvisioned ||
    []

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Cluster recommendations</h1>
          <p>Right-sizing, pressure, and capacity signals for cluster {id}.</p>
        </div>

        <div className="cluster-actions">
          <button
            type="button"
            className="dashboard-nav-button"
            onClick={() => navigate('/dashboard/clusters')}
          >
            Back to clusters
          </button>
        </div>
      </div>

      <section className="dashboard-cards">
        <div className="dashboard-card">
          <h3>Analysis window</h3>

          <div className="form-grid">
            <label>
              <span>From</span>
              <input
                type="datetime-local"
                value={from}
                onChange={(e) => setFrom(e.target.value)}
              />
            </label>

            <label>
              <span>To</span>
              <input
                type="datetime-local"
                value={to}
                onChange={(e) => setTo(e.target.value)}
              />
            </label>
          </div>

          <div className="cluster-actions">
            <button
              type="button"
              className="dashboard-nav-button is-primary"
              onClick={loadData}
              disabled={isLoading}
            >
              Refresh
            </button>
          </div>

          {error && <div className="form-error">{error}</div>}
        </div>
      </section>

      <section className="dashboard-cards">
        <div className="dashboard-card">
          <h3>Capacity overview</h3>

          {isLoading && !capacity ? (
            <p>Loading capacity...</p>
          ) : !capacity ? (
            <p>No capacity data available.</p>
          ) : (
            <div className="agents-list">
              <div className="agent-list-item">
                <div className="agent-list-main">
                  <h4>CPU</h4>
                  <p>Total: {formatNumber(capacity.totalCpuCores ?? capacity.cpuTotalCores)}</p>
                  <p>Requested: {formatNumber(capacity.requestedCpuCores ?? capacity.cpuRequestedCores)}</p>
                  <p>Recommended: {formatNumber(capacity.recommendedCpuCores ?? capacity.cpuRecommendedCores)}</p>
                  <p>Headroom: {formatNumber(capacity.cpuHeadroomCores ?? capacity.availableCpuCores)}</p>
                </div>
              </div>

              <div className="agent-list-item">
                <div className="agent-list-main">
                  <h4>Memory</h4>
                  <p>Total: {formatBytes(capacity.totalMemoryBytes ?? capacity.memoryTotalBytes)}</p>
                  <p>Requested: {formatBytes(capacity.requestedMemoryBytes ?? capacity.memoryRequestedBytes)}</p>
                  <p>Recommended: {formatBytes(capacity.recommendedMemoryBytes ?? capacity.memoryRecommendedBytes)}</p>
                  <p>Headroom: {formatBytes(capacity.memoryHeadroomBytes ?? capacity.availableMemoryBytes)}</p>
                </div>
              </div>
            </div>
          )}
        </div>
      </section>

      <section className="dashboard-cards">
        <div className="dashboard-card">
          <h3>Right-sizing recommendations</h3>

          {isLoading && !recommendationItems.length ? (
            <p>Loading recommendations...</p>
          ) : recommendationItems.length === 0 ? (
            <p>No right-sizing recommendations found for the selected period.</p>
          ) : (
            <div className="agents-list">
              {recommendationItems.map((item, index) => {
                const key =
                  item.workloadUid ||
                  item.workloadName ||
                  item.name ||
                  `${item.namespace || 'ns'}-${index}`

                return (
                  <div key={key} className="agent-list-item">
                    <div className="agent-list-main">
                      <h4>{item.workloadName || item.name || 'Unnamed workload'}</h4>
                      <p>Namespace: {item.namespace || '—'}</p>
                      <p>Kind: {item.kind || '—'}</p>
                      <p>Replicas: {item.replicas ?? '—'}</p>
                      <p>Current CPU request: {formatNumber(item.currentCpuRequestCores)}</p>
                      <p>Recommended CPU request: {formatNumber(item.recommendedCpuRequestCores)}</p>
                      <p>Current memory request: {formatBytes(item.currentMemoryRequestBytes)}</p>
                      <p>Recommended memory request: {formatBytes(item.recommendedMemoryRequestBytes)}</p>
                    </div>

                    <div className="agent-list-side">
                      {item.savingsCpuCores !== undefined && (
                        <p>CPU delta: {formatNumber(item.savingsCpuCores)}</p>
                      )}
                      {item.savingsMemoryBytes !== undefined && (
                        <p>Memory delta: {formatBytes(item.savingsMemoryBytes)}</p>
                      )}
                      {item.reason && <p>{item.reason}</p>}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </section>

      <section className="dashboard-cards">
        <div className="dashboard-card">
          <h3>Over-provisioned workloads</h3>

          {isLoading && !Array.isArray(overProvisionedItems) ? (
            <p>Loading over-provisioned workloads...</p>
          ) : !Array.isArray(overProvisionedItems) || overProvisionedItems.length === 0 ? (
            <p>No over-provisioned workloads detected.</p>
          ) : (
            <div className="agents-list">
              {overProvisionedItems.map((item, index) => {
                const key =
                  item.workloadUid ||
                  item.workloadName ||
                  item.name ||
                  `over-${index}`

                return (
                  <div key={key} className="agent-list-item">
                    <div className="agent-list-main">
                      <h4>{item.workloadName || item.name || 'Unnamed workload'}</h4>
                      <p>Namespace: {item.namespace || '—'}</p>
                      <p>Kind: {item.kind || '—'}</p>
                      <p>CPU request: {formatNumber(item.cpuRequestCores)}</p>
                      <p>CPU recommended: {formatNumber(item.recommendedCpuRequestCores)}</p>
                      <p>Memory request: {formatBytes(item.memoryRequestBytes)}</p>
                      <p>Memory recommended: {formatBytes(item.recommendedMemoryRequestBytes)}</p>
                    </div>

                    <div className="agent-list-side">
                      {item.reclaimCpuCores !== undefined && (
                        <p>Reclaim CPU: {formatNumber(item.reclaimCpuCores)}</p>
                      )}
                      {item.reclaimMemoryBytes !== undefined && (
                        <p>Reclaim memory: {formatBytes(item.reclaimMemoryBytes)}</p>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </section>

      <section className="dashboard-cards">
        <div className="dashboard-card">
          <h3>Under-provisioned workloads</h3>

          {isLoading && !Array.isArray(underProvisionedItems) ? (
            <p>Loading under-provisioned workloads...</p>
          ) : !Array.isArray(underProvisionedItems) || underProvisionedItems.length === 0 ? (
            <p>No under-provisioned workloads detected.</p>
          ) : (
            <div className="agents-list">
              {underProvisionedItems.map((item, index) => {
                const key =
                  item.workloadUid ||
                  item.workloadName ||
                  item.name ||
                  `under-${index}`

                return (
                  <div key={key} className="agent-list-item">
                    <div className="agent-list-main">
                      <h4>{item.workloadName || item.name || 'Unnamed workload'}</h4>
                      <p>Namespace: {item.namespace || '—'}</p>
                      <p>Kind: {item.kind || '—'}</p>
                      <p>CPU pressure frequency: {formatNumber(item.cpuPressureFrequency, 3)}</p>
                      <p>Memory pressure frequency: {formatNumber(item.memoryPressureFrequency, 3)}</p>
                      <p>Peak CPU usage: {formatNumber(item.peakCpuUsageCores)}</p>
                      <p>Peak memory usage: {formatBytes(item.peakMemoryUsageBytes)}</p>
                    </div>

                    <div className="agent-list-side">
                      {item.reason && <p>{item.reason}</p>}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </section>
    </>
  )
}