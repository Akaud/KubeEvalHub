import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  BarChart,
  Bar,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
} from 'recharts'

function formatXAxis(value) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value

  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}:${String(date.getSeconds()).padStart(2, '0')}`
}

function buildSeriesTitle(series) {
  const segments = [
    series.metricName,
    series.resourceKind,
    series.nodeName,
    series.namespace,
    series.podName,
    series.containerName,
  ].filter(Boolean)

  return segments.join(' / ')
}

function getYAxisDomain(samples) {
  if (!samples || samples.length === 0) {
    return ['auto', 'auto']
  }

  const values = samples
    .map((sample) => sample.value)
    .filter((value) => Number.isFinite(value))

  if (values.length === 0) {
    return ['auto', 'auto']
  }

  const min = Math.min(...values)
  const max = Math.max(...values)

  if (min === max) {
    const padding = min === 0 ? 1 : Math.abs(min) * 0.2
    return [min - padding, max + padding]
  }

  const range = max - min
  const padding = range * 0.15

  return [min - padding, max + padding]
}

export default function ClusterMetricsDetailPage() {
  const { agentId } = useParams()
  const navigate = useNavigate()
  const { token } = useAuth()

  const [cluster, setCluster] = useState(null)
  const [metrics, setMetrics] = useState(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [error, setError] = useState('')
  const [selectedSeriesId, setSelectedSeriesId] = useState('')
  const [chartType, setChartType] = useState('line')

  const authToken = useMemo(() => {
    return token || localStorage.getItem('token') || ''
  }, [token])

  const loadClusterAndMetrics = useCallback(async (refresh = false) => {
    if (!authToken || !agentId) return

    if (refresh) {
      setIsRefreshing(true)
    } else {
      setIsLoading(true)
    }

    setError('')

    try {
      const clustersRes = await fetch('/api/clusters', {
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      })

      const clustersData = await clustersRes.json().catch(() => null)

      if (!clustersRes.ok) {
        throw new Error(clustersData?.error || 'Failed to load clusters')
      }

      const clusters = Array.isArray(clustersData) ? clustersData : []
      const currentCluster = clusters.find((item) => item.agentId === agentId) || null
      setCluster(currentCluster)

      const metricsRes = await fetch(
        `/api/clusters/${agentId}/metrics?from=2000-01-01T00:00:00Z&to=2100-01-01T00:00:00Z`,
        {
          headers: {
            Authorization: `Bearer ${authToken}`,
          },
        }
      )

      const metricsData = await metricsRes.json().catch(() => null)

      if (!metricsRes.ok) {
        throw new Error(metricsData?.error || 'Failed to load metrics')
      }

      setMetrics(metricsData)
    } catch (err) {
      setError(err.message || 'Failed to load metrics')
    } finally {
      setIsLoading(false)
      setIsRefreshing(false)
    }
  }, [authToken, agentId])

  useEffect(() => {
    loadClusterAndMetrics(false)
  }, [loadClusterAndMetrics])

  useEffect(() => {
    if (!metrics?.items?.length) {
      setSelectedSeriesId('')
      return
    }

    const stillExists = metrics.items.some((item) => item.series.id === selectedSeriesId)
    if (!selectedSeriesId || !stillExists) {
      setSelectedSeriesId(metrics.items[0].series.id)
    }
  }, [metrics, selectedSeriesId])

  const selectedItem = useMemo(() => {
    return metrics?.items?.find((item) => item.series.id === selectedSeriesId) || null
  }, [metrics, selectedSeriesId])

  const yAxisDomain = useMemo(() => {
    return getYAxisDomain(selectedItem?.samples || [])
  }, [selectedItem])

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>{cluster?.clusterName || 'Cluster metrics'}</h1>
          <p>Metric graphs for the selected cluster.</p>
        </div>
      </div>

      <section className="dashboard-cards dashboard-cards-single">
        <div className="dashboard-card">
          <div className="metrics-page-toolbar">
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
              onClick={() => loadClusterAndMetrics(true)}
              disabled={isLoading || isRefreshing}
            >
              {isRefreshing ? 'Refreshing...' : 'Refresh metrics'}
            </button>
          </div>

          {error && <div className="form-error">{error}</div>}

          {isLoading ? (
            <p>Loading metrics...</p>
          ) : !metrics?.items?.length ? (
            <p>No metrics available.</p>
          ) : (
            <>
              <div className="metrics-controls">
                <div className="metrics-control-group">
                  <label htmlFor="metric-select">Metric</label>
                  <select
                    id="metric-select"
                    value={selectedSeriesId}
                    onChange={(e) => setSelectedSeriesId(e.target.value)}
                  >
                    {metrics.items.map((item) => (
                      <option key={item.series.id} value={item.series.id}>
                        {buildSeriesTitle(item.series)}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="metrics-control-group">
                  <label htmlFor="chart-type-select">Chart type</label>
                  <select
                    id="chart-type-select"
                    value={chartType}
                    onChange={(e) => setChartType(e.target.value)}
                  >
                    <option value="line">Line</option>
                    <option value="bar">Bar</option>
                  </select>
                </div>
              </div>

              {selectedItem && (
                <div className="metrics-chart-card">
                  <h3>{buildSeriesTitle(selectedItem.series)}</h3>
                  <p>
                    Type: {selectedItem.series.metricType} | Unit: {selectedItem.series.unit}
                  </p>

                  <div className="metrics-chart-wrapper">
                    <ResponsiveContainer width="100%" height={360}>
                      {chartType === 'line' ? (
                        <AreaChart data={selectedItem.samples}>
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis
                            dataKey="collectedAt"
                            tickFormatter={formatXAxis}
                            minTickGap={24}
                          />
                          <YAxis domain={yAxisDomain} />
                          <Tooltip
                            labelFormatter={(label) =>
                              new Date(label).toLocaleString()
                            }
                          />
                          <Area
                            type="monotone"
                            dataKey="value"
                            stroke="#60a5fa"
                            fill="#60a5fa"
                            fillOpacity={0.25}
                            strokeWidth={2}
                            dot={false}
                          />
                        </AreaChart>
                      ) : (
                        <BarChart data={selectedItem.samples}>
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis
                            dataKey="collectedAt"
                            tickFormatter={formatXAxis}
                            minTickGap={24}
                          />
                          <YAxis domain={yAxisDomain} />
                          <Tooltip
                            labelFormatter={(label) =>
                              new Date(label).toLocaleString()
                            }
                          />
                          <Bar dataKey="value" fill="#60a5fa" />
                        </BarChart>
                      )}
                    </ResponsiveContainer>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </section>
    </>
  )
}