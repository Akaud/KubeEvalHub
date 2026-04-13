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

function clampFromToLast24Hours(fromValue, toValue) {
  if (!fromValue || !toValue) return fromValue

  const fromDate = new Date(fromValue)
  const toDate = new Date(toValue)

  if (Number.isNaN(fromDate.getTime()) || Number.isNaN(toDate.getTime())) {
    return fromValue
  }

  const minFrom = new Date(toDate)
  minFrom.setHours(minFrom.getHours() - 24)

  if (fromDate < minFrom) {
    return formatDateTimeInput(minFrom)
  }

  return fromValue
}

async function readJsonSafely(res) {
  const text = await res.text()

  try {
    return JSON.parse(text)
  } catch {
    throw new Error(`Expected JSON but got: ${text.slice(0, 200)}`)
  }
}

export default function ClusterRecommendationsPage() {
  const { agentId } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated, isReady } = useAuth()

  const defaultTo = useMemo(() => new Date(), [])
  const defaultFrom = useMemo(() => {
    const d = new Date()
    d.setHours(d.getHours() - 24)
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
    if (!isAuthenticated || !agentId) return

    setIsLoading(true)
    setError('')

    try {
      const query = buildQuery({
        from: toRFC3339Local(from),
        to: toRFC3339Local(to),
      })

      const requests = [
        {
          url: `/api/clusters/${agentId}/recommendations?${query}`,
          setter: setRecommendations,
          errorMessage: 'Failed to load recommendations',
        },
        {
          url: `/api/clusters/${agentId}/analysis/overprovisioned?${query}`,
          setter: setOverProvisioned,
          errorMessage: 'Failed to load over-provisioned workloads',
        },
        {
          url: `/api/clusters/${agentId}/analysis/underprovisioned?${query}`,
          setter: setUnderProvisioned,
          errorMessage: 'Failed to load under-provisioned workloads',
        },
        {
          url: `/api/clusters/${agentId}/capacity?${query}`,
          setter: setCapacity,
          errorMessage: 'Failed to load capacity',
        },
      ]

      const results = await Promise.allSettled(
        requests.map(async ({ url, setter, errorMessage }) => {
          const res = await apiFetch(url)
          const data = await readJsonSafely(res)

          if (!res.ok) {
            throw new Error(data?.error || errorMessage)
          }

          setter(data)
          return data
        })
      )

      const failures = results
        .filter((result) => result.status === 'rejected')
        .map((result) => result.reason?.message)
        .filter(Boolean)

      if (failures.length > 0) {
        setError(failures[0])
      }
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
  }, [isReady, isAuthenticated, agentId, from, to])

  const recommendationItems = Array.isArray(recommendations)
    ? recommendations
    : recommendations?.items ||
      recommendations?.recommendations ||
      recommendations?.workloads ||
      []

  const overProvisionedItems = Array.isArray(overProvisioned)
    ? overProvisioned
    : overProvisioned?.items ||
      overProvisioned?.workloads ||
      overProvisioned?.overProvisioned ||
      []

  const underProvisionedItems = Array.isArray(underProvisioned)
    ? underProvisioned
    : underProvisioned?.items ||
      underProvisioned?.workloads ||
      underProvisioned?.underProvisioned ||
      []

  const maxToInput = formatDateTimeInput(new Date())

  const minFromInput = useMemo(() => {
    const d = new Date(to || new Date())
    d.setHours(d.getHours() - 24)
    return formatDateTimeInput(d)
  }, [to])

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Cluster recommendations</h1>
          <p>Right-sizing, pressure, and capacity signals for cluster {agentId}.</p>
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
                min={minFromInput}
                max={to}
                onChange={(e) => {
                  const nextFrom = clampFromToLast24Hours(e.target.value, to)
                  setFrom(nextFrom)
                }}
              />
            </label>

            <label>
              <span>To</span>
              <input
                type="datetime-local"
                value={to}
                max={maxToInput}
                onChange={(e) => {
                  const nextTo = e.target.value
                  setTo(nextTo)
                  setFrom((currentFrom) => clampFromToLast24Hours(currentFrom, nextTo))
                }}
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
                  <p>Requested: {formatNumber(capacity.totalCpuRequestCores)}</p>
                  <p>Reclaimable: {formatNumber(capacity.reclaimableCpuCores ?? capacity.reclaimableCPUCores)}</p>
                  <p>Eligible workloads: {capacity.eligibleWorkloads ?? '—'}</p>
                  <p>Total workloads: {capacity.totalWorkloads ?? '—'}</p>
                </div>
              </div>

              <div className="agent-list-item">
                <div className="agent-list-main">
                  <h4>Memory</h4>
                  <p>Requested: {formatBytes(capacity.totalMemoryRequestBytes)}</p>
                  <p>Reclaimable: {formatBytes(capacity.reclaimableMemoryBytes)}</p>
                </div>
              </div>
            </div>
          )}
        </div>
      </section>

      <section className="dashboard-cards">
        <div className="dashboard-card">
          <h3>Right-sizing recommendations</h3>

          {isLoading && recommendations === null ? (
            <p>Loading recommendations...</p>
          ) : recommendationItems.length === 0 ? (
            <p>No right-sizing recommendations found for the selected period.</p>
          ) : (
            <div className="agents-list">
              {recommendationItems.map((item, index) => {
                const key =
                  item.workloadUid ||
                  item.workloadName ||
                  item.controllerUID ||
                  item.controllerName ||
                  item.name ||
                  `${item.namespace || 'ns'}-${index}`

                return (
                  <div key={key} className="agent-list-item">
                    <div className="agent-list-main">
                      <h4>{item.workloadName || item.controllerName || item.name || 'Unnamed workload'}</h4>
                      <p>Namespace: {item.namespace || '—'}</p>
                      <p>Kind: {item.kind || item.controllerKind || '—'}</p>
                      <p>Replicas: {item.replicas ?? '—'}</p>
                      <p>Current CPU request: {formatNumber(item.currentCpuRequestCores ?? item.cpuRequestCores)}</p>
                      <p>Recommended CPU request: {formatNumber(item.recommendedCpuRequestCores)}</p>
                      <p>Current memory request: {formatBytes(item.currentMemoryRequestBytes ?? item.memoryRequestBytes)}</p>
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

          {isLoading && overProvisioned === null ? (
            <p>Loading over-provisioned workloads...</p>
          ) : !Array.isArray(overProvisionedItems) || overProvisionedItems.length === 0 ? (
            <p>No over-provisioned workloads detected.</p>
          ) : (
            <div className="agents-list">
              {overProvisionedItems.map((item, index) => {
                const key =
                  item.workloadUid ||
                  item.workloadName ||
                  item.controllerUID ||
                  item.controllerName ||
                  item.name ||
                  `over-${index}`

                return (
                  <div key={key} className="agent-list-item">
                    <div className="agent-list-main">
                      <h4>{item.workloadName || item.controllerName || item.name || 'Unnamed workload'}</h4>
                      <p>Namespace: {item.namespace || '—'}</p>
                      <p>Kind: {item.kind || item.controllerKind || '—'}</p>
                      <p>CPU request: {formatNumber(item.cpuRequestCores ?? item.currentCpuRequestCores)}</p>
                      <p>CPU recommended: {formatNumber(item.recommendedCpuRequestCores ?? item.recommendedCPURequestCores)}</p>
                      <p>CPU p95: {formatNumber(item.observedCpuP95Cores ?? item.observedCPUP95Cores)}</p>
                      <p>Memory request: {formatBytes(item.memoryRequestBytes ?? item.currentMemoryRequestBytes)}</p>
                      <p>Memory recommended: {formatBytes(item.recommendedMemoryRequestBytes)}</p>
                      <p>Memory p95: {formatBytes(item.observedMemoryP95Bytes)}</p>
                    </div>

                    <div className="agent-list-side">
                      {(item.reclaimCpuCores !== undefined || item.reclaimableCpuCores !== undefined || item.reclaimableCPUCores !== undefined) && (
                        <p>
                          Reclaim CPU: {formatNumber(
                            item.reclaimCpuCores ??
                            item.reclaimableCpuCores ??
                            item.reclaimableCPUCores
                          )}
                        </p>
                      )}

                      {(item.reclaimMemoryBytes !== undefined || item.reclaimableMemoryBytes !== undefined) && (
                        <p>
                          Reclaim memory: {formatBytes(
                            item.reclaimMemoryBytes ??
                            item.reclaimableMemoryBytes
                          )}
                        </p>
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
          <h3>Under-provisioned workloads</h3>

          {isLoading && underProvisioned === null ? (
            <p>Loading under-provisioned workloads...</p>
          ) : !Array.isArray(underProvisionedItems) || underProvisionedItems.length === 0 ? (
            <p>No under-provisioned workloads detected.</p>
          ) : (
            <div className="agents-list">
              {underProvisionedItems.map((item, index) => {
                const key =
                  item.workloadUid ||
                  item.workloadName ||
                  item.controllerUID ||
                  item.controllerName ||
                  item.name ||
                  `under-${index}`

                return (
                  <div key={key} className="agent-list-item">
                    <div className="agent-list-main">
                      <h4>{item.workloadName || item.controllerName || item.name || 'Unnamed workload'}</h4>
                      <p>Namespace: {item.namespace || '—'}</p>
                      <p>Kind: {item.kind || item.controllerKind || '—'}</p>
                      <p>CPU pressure frequency: {formatNumber(item.cpuPressureFrequency, 3)}</p>
                      <p>Memory pressure frequency: {formatNumber(item.memoryPressureFrequency, 3)}</p>
                      <p>Peak CPU usage: {formatNumber(item.peakCpuUsageCores ?? item.observedCpuP95Cores ?? item.observedCPUP95Cores)}</p>
                      <p>Peak memory usage: {formatBytes(item.peakMemoryUsageBytes ?? item.observedMemoryP95Bytes)}</p>
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