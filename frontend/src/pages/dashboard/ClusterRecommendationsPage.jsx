import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { apiFetch } from '../../utils/apiFetch'
import { useAuth } from '../../context/AuthContext'
import '../../styles/ClusterRecommendationsPage.css'

function formatDateInput(date) {
  const pad = (n) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function formatTimeInput(date) {
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function buildDateTimeLocal(dateValue, timeValue) {
  if (!dateValue || !timeValue) return ''
  return `${dateValue}T${timeValue}`
}

function isCompleteDateTimeLocal(value) {
  return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(value)
}

function parseDateTimeLocal(value) {
  if (!isCompleteDateTimeLocal(value)) return null

  const [datePart, timePart] = value.split('T')
  const [year, month, day] = datePart.split('-').map(Number)
  const [hours, minutes] = timePart.split(':').map(Number)

  const parsed = new Date(year, month - 1, day, hours, minutes)
  return Number.isNaN(parsed.getTime()) ? null : parsed
}

function toRFC3339Local(value) {
  const parsed = parseDateTimeLocal(value)
  return parsed ? parsed.toISOString() : ''
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
  if (!isCompleteDateTimeLocal(fromValue) || !isCompleteDateTimeLocal(toValue)) {
    return fromValue
  }

  const fromDate = parseDateTimeLocal(fromValue)
  const toDate = parseDateTimeLocal(toValue)

  if (!fromDate || !toDate) {
    return fromValue
  }

  const minFrom = new Date(toDate)
  minFrom.setHours(minFrom.getHours() - 24)

  if (fromDate < minFrom) {
    return `${formatDateInput(minFrom)}T${formatTimeInput(minFrom)}`
  }

  if (fromDate > toDate) {
    return `${formatDateInput(toDate)}T${formatTimeInput(toDate)}`
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

function MetricRow({ label, value, highlight = false }) {
  return (
    <div className={`recommendation-metric-row${highlight ? ' is-highlighted' : ''}`}>
      <span className="recommendation-metric-label">{label}</span>
      <strong className="recommendation-metric-value">{value}</strong>
    </div>
  )
}

export default function ClusterRecommendationsPage() {
  const { clusterId } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated, isReady } = useAuth()

  const defaultTo = useMemo(() => new Date(), [])
  const defaultFrom = useMemo(() => {
    const d = new Date()
    d.setHours(d.getHours() - 24)
    return d
  }, [])

  const [fromDate, setFromDate] = useState(formatDateInput(defaultFrom))
  const [fromTime, setFromTime] = useState(formatTimeInput(defaultFrom))
  const [toDate, setToDate] = useState(formatDateInput(defaultTo))
  const [toTime, setToTime] = useState(formatTimeInput(defaultTo))

  const from = useMemo(
    () => buildDateTimeLocal(fromDate, fromTime),
    [fromDate, fromTime]
  )

  const to = useMemo(
    () => buildDateTimeLocal(toDate, toTime),
    [toDate, toTime]
  )

  const [recommendations, setRecommendations] = useState(null)
  const [overProvisioned, setOverProvisioned] = useState(null)
  const [underProvisioned, setUnderProvisioned] = useState(null)
  const [capacity, setCapacity] = useState(null)

  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const loadData = async () => {
    if (!isAuthenticated || !clusterId) return

    if (!isCompleteDateTimeLocal(from) || !isCompleteDateTimeLocal(to)) {
      setError('from and to are required')
      return
    }

    setIsLoading(true)
    setError('')

    try {
      const query = buildQuery({
        from: toRFC3339Local(from),
        to: toRFC3339Local(to),
      })

      const requests = [
        {
          url: `/api/clusters/${clusterId}/recommendations?${query}`,
          setter: setRecommendations,
          errorMessage: 'Failed to load recommendations',
        },
        {
          url: `/api/clusters/${clusterId}/analysis/overprovisioned?${query}`,
          setter: setOverProvisioned,
          errorMessage: 'Failed to load over-provisioned workloads',
        },
        {
          url: `/api/clusters/${clusterId}/analysis/underprovisioned?${query}`,
          setter: setUnderProvisioned,
          errorMessage: 'Failed to load under-provisioned workloads',
        },
        {
          url: `/api/clusters/${clusterId}/capacity?${query}`,
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
    if (!isCompleteDateTimeLocal(from) || !isCompleteDateTimeLocal(to)) return

    loadData()
    const intervalId = setInterval(loadData, 30000)

    return () => clearInterval(intervalId)
  }, [isReady, isAuthenticated, clusterId, from, to])

  useEffect(() => {
    if (!isCompleteDateTimeLocal(from) || !isCompleteDateTimeLocal(to)) return

    const clampedFrom = clampFromToLast24Hours(from, to)
    if (clampedFrom !== from) {
      const parsed = parseDateTimeLocal(clampedFrom)
      if (parsed) {
        setFromDate(formatDateInput(parsed))
        setFromTime(formatTimeInput(parsed))
      }
    }
  }, [from, to])

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

  const maxToDate = formatDateInput(new Date())
  const maxToTime = useMemo(() => {
    if (toDate !== maxToDate) return '23:59'
    return formatTimeInput(new Date())
  }, [toDate, maxToDate])

  const minFromDate = useMemo(() => {
    if (!isCompleteDateTimeLocal(to)) return ''
    const parsedTo = parseDateTimeLocal(to)
    if (!parsedTo) return ''

    const min = new Date(parsedTo)
    min.setHours(min.getHours() - 24)
    return formatDateInput(min)
  }, [to])

  const minFromTime = useMemo(() => {
    if (!isCompleteDateTimeLocal(to) || !fromDate) return '00:00'

    const parsedTo = parseDateTimeLocal(to)
    if (!parsedTo) return '00:00'

    const min = new Date(parsedTo)
    min.setHours(min.getHours() - 24)

    if (fromDate !== formatDateInput(min)) return '00:00'
    return formatTimeInput(min)
  }, [to, fromDate])

  const maxFromTime = useMemo(() => {
    if (!isCompleteDateTimeLocal(to) || !fromDate || fromDate !== toDate) {
      return '23:59'
    }
    return toTime || '23:59'
  }, [to, fromDate, toDate, toTime])

  const applyClampedFrom = (nextDate, nextTime) => {
    const nextFrom = buildDateTimeLocal(nextDate, nextTime)

    if (!nextDate) {
      setFromDate('')
      return
    }

    if (!nextTime) {
      setFromTime('')
      return
    }

    if (!isCompleteDateTimeLocal(nextFrom) || !isCompleteDateTimeLocal(to)) {
      setFromDate(nextDate)
      setFromTime(nextTime)
      return
    }

    const clamped = clampFromToLast24Hours(nextFrom, to)
    const parsed = parseDateTimeLocal(clamped)

    if (!parsed) {
      setFromDate(nextDate)
      setFromTime(nextTime)
      return
    }

    setFromDate(formatDateInput(parsed))
    setFromTime(formatTimeInput(parsed))
  }

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Cluster recommendations</h1>
          <p>Right-sizing, pressure, and capacity signals for cluster {clusterId}.</p>
        </div>

        <div className="cluster-actions">
          <button
            type="button"
            className="dashboard-nav-button"
            onClick={() => navigate('/dashboard')}
          >
            Back to clusters
          </button>
        </div>
      </div>

      <div className="cluster-recommendations-page">
        <section className="cluster-recommendations-top">
          <div className="cluster-recommendations-main">
            <div className="dashboard-card cluster-recommendations-card analysis-window-card">
              <h3>Analysis window</h3>

              <div className="analysis-window-grid">
                <div className="analysis-window-field">
                  <span>From</span>
                  <div className="analysis-window-split">
                    <input
                      className="analysis-window-input"
                      type="date"
                      value={fromDate}
                      min={minFromDate || undefined}
                      max={toDate || undefined}
                      onChange={(e) => {
                        const nextDate = e.target.value
                        setFromDate(nextDate)
                        applyClampedFrom(nextDate, fromTime)
                      }}
                    />
                    <input
                      className="analysis-window-input"
                      type="time"
                      step={60}
                      value={fromTime}
                      min={minFromTime}
                      max={maxFromTime}
                      onChange={(e) => {
                        const nextTime = e.target.value
                        setFromTime(nextTime)
                        applyClampedFrom(fromDate, nextTime)
                      }}
                    />
                  </div>
                </div>

                <div className="analysis-window-field">
                  <span>To</span>
                  <div className="analysis-window-split">
                    <input
                      className="analysis-window-input"
                      type="date"
                      value={toDate}
                      max={maxToDate}
                      onChange={(e) => {
                        const nextDate = e.target.value
                        setToDate(nextDate)
                      }}
                    />
                    <input
                      className="analysis-window-input"
                      type="time"
                      step={60}
                      value={toTime}
                      min="00:00"
                      max={toDate === maxToDate ? maxToTime : '23:59'}
                      onChange={(e) => {
                        const nextTime = e.target.value
                        setToTime(nextTime)
                      }}
                    />
                  </div>
                </div>
              </div>

              <div className="analysis-window-actions">
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
          </div>

          <aside className="cluster-recommendations-sidebar">
            <div className="dashboard-card cluster-recommendations-card capacity-card">
              <h3>Capacity overview</h3>

              {isLoading && !capacity ? (
                <p>Loading capacity...</p>
              ) : !capacity ? (
                <p>No capacity data available.</p>
              ) : (
                <div className="capacity-grid">
                  <div className="capacity-stat-card">
                    <h4>CPU</h4>
                    <div className="capacity-stat-list">
                      <div className="capacity-stat-row">
                        <span>Requested</span>
                        <strong>{formatNumber(capacity.totalCpuRequestCores)}</strong>
                      </div>
                      <div className="capacity-stat-row">
                        <span>Reclaimable</span>
                        <strong>
                          {formatNumber(
                            capacity.reclaimableCpuCores ?? capacity.reclaimableCPUCores
                          )}
                        </strong>
                      </div>
                      <div className="capacity-stat-row">
                        <span>Eligible workloads</span>
                        <strong>{capacity.eligibleWorkloads ?? '—'}</strong>
                      </div>
                      <div className="capacity-stat-row">
                        <span>Total workloads</span>
                        <strong>{capacity.totalWorkloads ?? '—'}</strong>
                      </div>
                    </div>
                  </div>

                  <div className="capacity-stat-card">
                    <h4>Memory</h4>
                    <div className="capacity-stat-list">
                      <div className="capacity-stat-row">
                        <span>Requested</span>
                        <strong>{formatBytes(capacity.totalMemoryRequestBytes)}</strong>
                      </div>
                      <div className="capacity-stat-row">
                        <span>Reclaimable</span>
                        <strong>{formatBytes(capacity.reclaimableMemoryBytes)}</strong>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </aside>
        </section>

        <section className="dashboard-cards dashboard-cards-single">
          <div className="dashboard-card recommendation-list-card">
            <h3>Right-sizing recommendations</h3>

            {isLoading && recommendations === null ? (
              <p>Loading recommendations...</p>
            ) : recommendationItems.length === 0 ? (
              <p>No right-sizing recommendations found for the selected period.</p>
            ) : (
              <div className="recommendation-grid">
                {recommendationItems.map((item, index) => {
                  const key =
                    item.workloadUid ||
                    item.workloadName ||
                    item.controllerUID ||
                    item.controllerName ||
                    item.name ||
                    `${item.namespace || 'ns'}-${index}`

                  return (
                    <div key={key} className="agent-list-item recommendation-grid-item">
                      <div className="agent-list-main">
                        <h4>{item.workloadName || item.controllerName || item.name || 'Unnamed workload'}</h4>

                        <div className="recommendation-meta">
                          <p>Namespace: {item.namespace || '—'}</p>
                          <p>Kind: {item.kind || item.controllerKind || '—'}</p>
                          <p>Replicas: {item.replicas ?? '—'}</p>
                        </div>

                        <div className="recommendation-metrics">
                          <MetricRow
                            label="Current CPU request"
                            value={formatNumber(item.currentCpuRequestCores ?? item.cpuRequestCores)}
                          />
                          <MetricRow
                            label="Recommended CPU request"
                            value={formatNumber(item.recommendedCpuRequestCores)}
                            highlight
                          />
                          <MetricRow
                            label="Current memory request"
                            value={formatBytes(item.currentMemoryRequestBytes ?? item.memoryRequestBytes)}
                          />
                          <MetricRow
                            label="Recommended memory request"
                            value={formatBytes(item.recommendedMemoryRequestBytes)}
                            highlight
                          />
                        </div>
                      </div>

                      <div className="agent-list-side">
                        {item.savingsCpuCores !== undefined && (
                          <MetricRow
                            label="CPU delta"
                            value={formatNumber(item.savingsCpuCores)}
                            highlight
                          />
                        )}
                        {item.savingsMemoryBytes !== undefined && (
                          <MetricRow
                            label="Memory delta"
                            value={formatBytes(item.savingsMemoryBytes)}
                            highlight
                          />
                        )}
                        {item.reason && <p className="recommendation-reason">{item.reason}</p>}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </section>

        <section className="dashboard-cards dashboard-cards-single">
          <div className="dashboard-card recommendation-list-card">
            <h3>Over-provisioned workloads</h3>

            {isLoading && overProvisioned === null ? (
              <p>Loading over-provisioned workloads...</p>
            ) : !Array.isArray(overProvisionedItems) || overProvisionedItems.length === 0 ? (
              <p>No over-provisioned workloads detected.</p>
            ) : (
              <div className="recommendation-grid">
                {overProvisionedItems.map((item, index) => {
                  const key =
                    item.workloadUid ||
                    item.workloadName ||
                    item.controllerUID ||
                    item.controllerName ||
                    item.name ||
                    `over-${index}`

                  return (
                    <div key={key} className="agent-list-item recommendation-grid-item">
                      <div className="agent-list-main">
                        <h4>{item.workloadName || item.controllerName || item.name || 'Unnamed workload'}</h4>

                        <div className="recommendation-meta">
                          <p>Namespace: {item.namespace || '—'}</p>
                          <p>Kind: {item.kind || item.controllerKind || '—'}</p>
                        </div>

                        <div className="recommendation-metrics">
                          <MetricRow
                            label="CPU request"
                            value={formatNumber(item.cpuRequestCores ?? item.currentCpuRequestCores)}
                          />
                          <MetricRow
                            label="CPU recommended"
                            value={formatNumber(item.recommendedCpuRequestCores ?? item.recommendedCPURequestCores)}
                            highlight
                          />
                          <MetricRow
                            label="CPU p95"
                            value={formatNumber(item.observedCpuP95Cores ?? item.observedCPUP95Cores)}
                          />
                          <MetricRow
                            label="Memory request"
                            value={formatBytes(item.memoryRequestBytes ?? item.currentMemoryRequestBytes)}
                          />
                          <MetricRow
                            label="Memory recommended"
                            value={formatBytes(item.recommendedMemoryRequestBytes)}
                            highlight
                          />
                          <MetricRow
                            label="Memory p95"
                            value={formatBytes(item.observedMemoryP95Bytes)}
                          />
                        </div>
                      </div>

                      <div className="agent-list-side">
                        {(item.reclaimCpuCores !== undefined ||
                          item.reclaimableCpuCores !== undefined ||
                          item.reclaimableCPUCores !== undefined) && (
                          <MetricRow
                            label="Reclaim CPU"
                            value={formatNumber(
                              item.reclaimCpuCores ??
                                item.reclaimableCpuCores ??
                                item.reclaimableCPUCores
                            )}
                            highlight
                          />
                        )}

                        {(item.reclaimMemoryBytes !== undefined ||
                          item.reclaimableMemoryBytes !== undefined) && (
                          <MetricRow
                            label="Reclaim memory"
                            value={formatBytes(
                              item.reclaimMemoryBytes ?? item.reclaimableMemoryBytes
                            )}
                            highlight
                          />
                        )}

                        {item.reason && <p className="recommendation-reason">{item.reason}</p>}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </section>

        <section className="dashboard-cards dashboard-cards-single">
          <div className="dashboard-card recommendation-list-card">
            <h3>Under-provisioned workloads</h3>

            {isLoading && underProvisioned === null ? (
              <p>Loading under-provisioned workloads...</p>
            ) : !Array.isArray(underProvisionedItems) || underProvisionedItems.length === 0 ? (
              <p>No under-provisioned workloads detected.</p>
            ) : (
              <div className="recommendation-grid">
                {underProvisionedItems.map((item, index) => {
                  const key =
                    item.workloadUid ||
                    item.workloadName ||
                    item.controllerUID ||
                    item.controllerName ||
                    item.name ||
                    `under-${index}`

                  return (
                    <div key={key} className="agent-list-item recommendation-grid-item">
                      <div className="agent-list-main">
                        <h4>{item.workloadName || item.controllerName || item.name || 'Unnamed workload'}</h4>

                        <div className="recommendation-meta">
                          <p>Namespace: {item.namespace || '—'}</p>
                          <p>Kind: {item.kind || item.controllerKind || '—'}</p>
                        </div>

                        <div className="recommendation-metrics">
                          <MetricRow
                            label="CPU pressure frequency"
                            value={formatNumber(item.cpuPressureFrequency, 3)}
                          />
                          <MetricRow
                            label="Memory pressure frequency"
                            value={formatNumber(item.memoryPressureFrequency, 3)}
                          />
                          <MetricRow
                            label="Peak CPU usage"
                            value={formatNumber(
                              item.peakCpuUsageCores ??
                                item.observedCpuP95Cores ??
                                item.observedCPUP95Cores
                            )}
                            highlight
                          />
                          <MetricRow
                            label="Peak memory usage"
                            value={formatBytes(
                              item.peakMemoryUsageBytes ?? item.observedMemoryP95Bytes
                            )}
                            highlight
                          />
                        </div>
                      </div>

                      <div className="agent-list-side">
                        {item.reason && <p className="recommendation-reason">{item.reason}</p>}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </section>
      </div>
    </>
  )
}