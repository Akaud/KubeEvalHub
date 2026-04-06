import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import {
  ResponsiveContainer,
  ComposedChart,
  Area,
  Line,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
  Brush,
} from 'recharts'

const NONE_KEY = '__none__'

const PREDICTION_MODELS = [
  {
    value: 'moving_average',
    label: 'Moving average',
    description: 'Uses the recent average level. Best for noisy metrics with no strong trend. Usually the safest default.',
  },
  {
    value: 'last_value',
    label: 'Last value',
    description: 'Assumes the next values will stay near the latest observed value. Best for stable or flat metrics.',
  },
  {
    value: 'damped_holt',
    label: 'Damped trend',
    description: 'Follows the recent trend, but gradually weakens it over time. Better than linear trend when growth or decline should not explode.',
  },
  {
    value: 'holt_linear',
    label: 'Linear trend',
    description: 'Projects the recent trend forward at full strength. Best only when the metric has a clear steady trend.',
  },
]

function formatXAxis(value) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value

  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}:${String(date.getSeconds()).padStart(2, '0')}`
}

function formatFullDateTime(value) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
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

function getYAxisDomain(samples, forecast = []) {
  const values = [
    ...(samples || []).map((sample) => sample.value),
    ...(forecast || []).map((point) => point.value),
  ].filter((value) => Number.isFinite(value))

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

function displayValue(value, fallback = 'N/A') {
  return value && value !== NONE_KEY ? value : fallback
}

function buildMetricsTree(items) {
  const tree = {}

  for (const item of items || []) {
    const s = item.series
    const metricName = s.metricName || NONE_KEY
    const resourceKind = s.resourceKind || NONE_KEY
    const scopeKey = s.resourceKind === 'node'
      ? (s.nodeName || NONE_KEY)
      : (s.namespace || NONE_KEY)
    const podKey = s.podName || NONE_KEY
    const containerKey = s.containerName || NONE_KEY

    if (!tree[metricName]) {
      tree[metricName] = {}
    }

    if (!tree[metricName][resourceKind]) {
      tree[metricName][resourceKind] = {}
    }

    if (!tree[metricName][resourceKind][scopeKey]) {
      tree[metricName][resourceKind][scopeKey] = {}
    }

    if (!tree[metricName][resourceKind][scopeKey][podKey]) {
      tree[metricName][resourceKind][scopeKey][podKey] = {}
    }

    tree[metricName][resourceKind][scopeKey][podKey][containerKey] = item
  }

  return tree
}

function getObjectKeys(obj) {
  return obj ? Object.keys(obj) : []
}

function buildChartData(samples, forecast) {
  const actualSamples = samples || []
  const forecastSamples = forecast || []

  const actual = actualSamples.map((sample, index) => ({
    collectedAt: sample.collectedAt,
    value: sample.value,
    forecastValue: index === actualSamples.length - 1 ? sample.value : null,
  }))

  if (forecastSamples.length === 0) {
    return actual
  }

  const predicted = forecastSamples.map((point) => ({
    collectedAt: point.collectedAt,
    value: null,
    forecastValue: point.value,
  }))

  return [...actual, ...predicted]
}

function MetricsTooltip({ active, payload, label }) {
  if (!active || !payload || !payload.length) {
    return null
  }

  const actual = payload.find((entry) => entry.dataKey === 'value' && entry.value != null)
  const forecast = payload.find((entry) => entry.dataKey === 'forecastValue' && entry.value != null)

  return (
    <div
      style={{
        background: '#166534',
        color: '#ffffff',
        border: '1px solid #15803d',
        borderRadius: '12px',
        padding: '12px 14px',
        boxShadow: '0 12px 24px rgba(0, 0, 0, 0.25)',
      }}
    >
      <div style={{ fontSize: '14px', fontWeight: 700, marginBottom: '6px' }}>
        {formatFullDateTime(label)}
      </div>

      {actual && (
        <div style={{ fontSize: '14px', marginBottom: forecast ? '4px' : 0 }}>
          Actual: {typeof actual.value === 'number' ? actual.value.toLocaleString() : actual.value}
        </div>
      )}

      {forecast && (
        <div style={{ fontSize: '14px' }}>
          Prediction: {typeof forecast.value === 'number' ? forecast.value.toLocaleString() : forecast.value}
        </div>
      )}
    </div>
  )
}

export default function ClusterMetricsDetailPage() {
  const { agentId } = useParams()
  const navigate = useNavigate()
  const { token } = useAuth()

  const [cluster, setCluster] = useState(null)
  const [metrics, setMetrics] = useState(null)
  const [forecast, setForecast] = useState(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [isPredicting, setIsPredicting] = useState(false)
  const [showPrediction, setShowPrediction] = useState(false)
  const [error, setError] = useState('')

  const [selectedMetricName, setSelectedMetricName] = useState('')
  const [selectedResourceKind, setSelectedResourceKind] = useState('')
  const [selectedScope, setSelectedScope] = useState('')
  const [selectedPod, setSelectedPod] = useState('')
  const [selectedContainer, setSelectedContainer] = useState('')

  const [predictionHistoryLimit, setPredictionHistoryLimit] = useState(120)
  const [predictionSteps, setPredictionSteps] = useState(8)
  const [predictionModel, setPredictionModel] = useState('moving_average')

  const authToken = useMemo(() => {
    return token || localStorage.getItem('token') || ''
  }, [token])

  const selectedPredictionModel = useMemo(() => {
    return PREDICTION_MODELS.find((model) => model.value === predictionModel) || PREDICTION_MODELS[0]
  }, [predictionModel])

  const loadClusterAndMetrics = useCallback(async (refresh = false) => {
    if (!authToken || !agentId) return

    if (refresh) {
      setIsRefreshing(true)
    } else {
      setIsLoading(true)
    }

    setError('')
    setForecast(null)
    setShowPrediction(false)

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

  const metricsTree = useMemo(() => {
    return buildMetricsTree(metrics?.items || [])
  }, [metrics])

  const metricNameOptions = useMemo(() => {
    return getObjectKeys(metricsTree).sort()
  }, [metricsTree])

  const resourceKindOptions = useMemo(() => {
    return getObjectKeys(metricsTree[selectedMetricName]).sort()
  }, [metricsTree, selectedMetricName])

  const scopeOptions = useMemo(() => {
    return getObjectKeys(metricsTree[selectedMetricName]?.[selectedResourceKind]).sort()
  }, [metricsTree, selectedMetricName, selectedResourceKind])

  const podOptions = useMemo(() => {
    return getObjectKeys(
      metricsTree[selectedMetricName]?.[selectedResourceKind]?.[selectedScope]
    ).sort()
  }, [metricsTree, selectedMetricName, selectedResourceKind, selectedScope])

  const containerOptions = useMemo(() => {
    return getObjectKeys(
      metricsTree[selectedMetricName]?.[selectedResourceKind]?.[selectedScope]?.[selectedPod]
    ).sort()
  }, [metricsTree, selectedMetricName, selectedResourceKind, selectedScope, selectedPod])

  useEffect(() => {
    if (!metricNameOptions.length) {
      setSelectedMetricName('')
      return
    }

    if (!metricNameOptions.includes(selectedMetricName)) {
      setSelectedMetricName(metricNameOptions[0])
    }
  }, [metricNameOptions, selectedMetricName])

  useEffect(() => {
    if (!resourceKindOptions.length) {
      setSelectedResourceKind('')
      return
    }

    if (!resourceKindOptions.includes(selectedResourceKind)) {
      setSelectedResourceKind(resourceKindOptions[0])
    }
  }, [resourceKindOptions, selectedResourceKind])

  useEffect(() => {
    if (!scopeOptions.length) {
      setSelectedScope('')
      return
    }

    if (!scopeOptions.includes(selectedScope)) {
      setSelectedScope(scopeOptions[0])
    }
  }, [scopeOptions, selectedScope])

  useEffect(() => {
    if (!podOptions.length) {
      setSelectedPod('')
      return
    }

    if (!podOptions.includes(selectedPod)) {
      setSelectedPod(podOptions[0])
    }
  }, [podOptions, selectedPod])

  useEffect(() => {
    if (!containerOptions.length) {
      setSelectedContainer('')
      return
    }

    if (!containerOptions.includes(selectedContainer)) {
      setSelectedContainer(containerOptions[0])
    }
  }, [containerOptions, selectedContainer])

  const selectedItem = useMemo(() => {
    return (
      metricsTree[selectedMetricName]?.[selectedResourceKind]?.[selectedScope]?.[selectedPod]?.[selectedContainer] ||
      null
    )
  }, [
    metricsTree,
    selectedMetricName,
    selectedResourceKind,
    selectedScope,
    selectedPod,
    selectedContainer,
  ])

  useEffect(() => {
    setForecast(null)
    setShowPrediction(false)
  }, [
    selectedMetricName,
    selectedResourceKind,
    selectedScope,
    selectedPod,
    selectedContainer,
  ])

  useEffect(() => {
    setForecast(null)
    setShowPrediction(false)
  }, [predictionHistoryLimit, predictionSteps, predictionModel])

  const loadPrediction = useCallback(async () => {
    if (!authToken || !agentId || !selectedItem) return

    setIsPredicting(true)
    setError('')

    try {
      const series = selectedItem.series
      const body = {
        seriesId: series.id,
        model: predictionModel,
        steps: predictionSteps,
        historyLimit: predictionHistoryLimit,
      }

      const res = await fetch(`/api/clusters/${agentId}/forecast`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      })

      const data = await res.json().catch(() => null)

      if (!res.ok) {
        throw new Error(data?.error || 'Failed to build prediction')
      }

      setForecast(data)
      setShowPrediction(true)
    } catch (err) {
      setError(err.message || 'Failed to build prediction')
      setForecast(null)
      setShowPrediction(false)
    } finally {
      setIsPredicting(false)
    }
  }, [authToken, agentId, selectedItem, predictionHistoryLimit, predictionSteps, predictionModel])

  const forecastPoints = useMemo(() => {
    if (!showPrediction) return []
    return Array.isArray(forecast?.forecast) ? forecast.forecast : []
  }, [forecast, showPrediction])

  const chartData = useMemo(() => {
    return buildChartData(selectedItem?.samples || [], forecastPoints)
  }, [selectedItem, forecastPoints])

  const yAxisDomain = useMemo(() => {
    return getYAxisDomain(selectedItem?.samples || [], forecastPoints)
  }, [selectedItem, forecastPoints])

  const showScopeSelector = useMemo(() => {
    if (!selectedResourceKind) return false
    return scopeOptions.length > 1 || (scopeOptions.length === 1 && scopeOptions[0] !== NONE_KEY)
  }, [selectedResourceKind, scopeOptions])

  const showPodSelector = useMemo(() => {
    return podOptions.length > 1 || (podOptions.length === 1 && podOptions[0] !== NONE_KEY)
  }, [podOptions])

  const showContainerSelector = useMemo(() => {
    return containerOptions.length > 1 || (containerOptions.length === 1 && containerOptions[0] !== NONE_KEY)
  }, [containerOptions])

  const scopeLabel = selectedResourceKind === 'node' ? 'Node' : 'Namespace'

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
              onClick={() => loadClusterAndMetrics(true)}
              disabled={isLoading || isRefreshing || isPredicting}
            >
              {isRefreshing ? 'Refreshing...' : 'Refresh metrics'}
            </button>
          </div>

          <div className="metrics-toolbar-right">
            <div className="metrics-toolbar-controls">
              <div className="metrics-toolbar-control">
                <label htmlFor="prediction-model-select">Model</label>
                <select
                  id="prediction-model-select"
                  value={predictionModel}
                  onChange={(e) => setPredictionModel(e.target.value)}
                  disabled={isLoading || isRefreshing || isPredicting}
                >
                  {PREDICTION_MODELS.map((model) => (
                    <option key={model.value} value={model.value}>
                      {model.label}
                    </option>
                  ))}
                </select>
              </div>

              <div className="metrics-toolbar-control">
                <label htmlFor="prediction-history-select">History</label>
                <select
                  id="prediction-history-select"
                  value={predictionHistoryLimit}
                  onChange={(e) => setPredictionHistoryLimit(Number(e.target.value))}
                  disabled={isLoading || isRefreshing || isPredicting}
                >
                  <option value={60}>60</option>
                  <option value={120}>120</option>
                  <option value={240}>240</option>
                  <option value={300}>300</option>
                </select>
              </div>

              <div className="metrics-toolbar-control">
                <label htmlFor="prediction-steps-select">Steps</label>
                <select
                  id="prediction-steps-select"
                  value={predictionSteps}
                  onChange={(e) => setPredictionSteps(Number(e.target.value))}
                  disabled={isLoading || isRefreshing || isPredicting}
                >
                  <option value={4}>4</option>
                  <option value={8}>8</option>
                  <option value={12}>12</option>
                  <option value={20}>20</option>
                </select>
              </div>
            </div>

            <button
              type="button"
              className={`dashboard-nav-button ${showPrediction ? 'is-primary' : ''}`}
              onClick={() => {
                if (showPrediction) {
                  setShowPrediction(false)
                } else {
                  loadPrediction()
                }
              }}
              disabled={!selectedItem || isLoading || isRefreshing || isPredicting}
            >
              {isPredicting ? 'Predicting...' : showPrediction ? 'Hide prediction' : 'Prediction'}
            </button>
          </div>
        </div>

          <div className="prediction-model-help">
            <strong>{selectedPredictionModel.label}</strong>
            <p>{selectedPredictionModel.description}</p>
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
                  <label htmlFor="metric-name-select">Metric</label>
                  <select
                    id="metric-name-select"
                    value={selectedMetricName}
                    onChange={(e) => setSelectedMetricName(e.target.value)}
                  >
                    {metricNameOptions.map((metricName) => (
                      <option key={metricName} value={metricName}>
                        {metricName}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="metrics-control-group">
                  <label htmlFor="resource-kind-select">Resource kind</label>
                  <select
                    id="resource-kind-select"
                    value={selectedResourceKind}
                    onChange={(e) => setSelectedResourceKind(e.target.value)}
                  >
                    {resourceKindOptions.map((resourceKind) => (
                      <option key={resourceKind} value={resourceKind}>
                        {resourceKind}
                      </option>
                    ))}
                  </select>
                </div>

                {showScopeSelector && (
                  <div className="metrics-control-group">
                    <label htmlFor="scope-select">{scopeLabel}</label>
                    <select
                      id="scope-select"
                      value={selectedScope}
                      onChange={(e) => setSelectedScope(e.target.value)}
                    >
                      {scopeOptions.map((scope) => (
                        <option key={scope} value={scope}>
                          {displayValue(scope)}
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                {showPodSelector && (
                  <div className="metrics-control-group">
                    <label htmlFor="pod-select">Pod</label>
                    <select
                      id="pod-select"
                      value={selectedPod}
                      onChange={(e) => setSelectedPod(e.target.value)}
                    >
                      {podOptions.map((pod) => (
                        <option key={pod} value={pod}>
                          {displayValue(pod)}
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                {showContainerSelector && (
                  <div className="metrics-control-group">
                    <label htmlFor="container-select">Container</label>
                    <select
                      id="container-select"
                      value={selectedContainer}
                      onChange={(e) => setSelectedContainer(e.target.value)}
                    >
                      {containerOptions.map((container) => (
                        <option key={container} value={container}>
                          {displayValue(container)}
                        </option>
                      ))}
                    </select>
                  </div>
                )}
              </div>

              {selectedItem && (
                <div className="metrics-chart-card">
                  <h3>{buildSeriesTitle(selectedItem.series)}</h3>
                  <p>
                    Type: {selectedItem.series.metricType} | Unit: {selectedItem.series.unit}
                    {showPrediction && forecast?.model?.name ? ` | Model: ${forecast.model.name}` : ''}
                  </p>

                  <div className="metrics-chart-wrapper">
                    <ResponsiveContainer width="100%" height={360}>
                      <ComposedChart data={chartData}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis
                          dataKey="collectedAt"
                          tickFormatter={formatXAxis}
                          minTickGap={24}
                        />
                        <YAxis domain={yAxisDomain} />
                        <Tooltip content={<MetricsTooltip />} />
                        <Area
                          type="monotone"
                          dataKey="value"
                          stroke="#60a5fa"
                          fill="#60a5fa"
                          fillOpacity={0.25}
                          strokeWidth={2}
                          dot={false}
                          isAnimationActive={false}
                          connectNulls={false}
                        />
                        {showPrediction && forecastPoints.length > 0 && (
                          <Line
                            type="monotone"
                            dataKey="forecastValue"
                            stroke="#22c55e"
                            strokeWidth={3}
                            dot={false}
                            strokeDasharray="8 6"
                            isAnimationActive={false}
                            connectNulls
                          />
                        )}
                        <Brush
                          dataKey="collectedAt"
                          height={28}
                          travellerWidth={10}
                          tickFormatter={formatXAxis}
                        />
                      </ComposedChart>
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