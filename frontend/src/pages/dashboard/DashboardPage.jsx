import { useEffect, useMemo, useState } from 'react'
import { NavLink } from 'react-router-dom'
import { FiBox, FiCpu, FiArrowRight } from 'react-icons/fi'
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
} from 'recharts'
import { apiFetch } from '../../utils/apiFetch'
import { useAuth } from '../../context/AuthContext'
import '../../styles/DashboardPage.css'

const HEARTBEAT_FRESH_MS = 10 * 60 * 1000

const CLUSTER_COLORS = {
  Active: '#22c55e',
  Inactive: '#f59e0b',
  'Never connected': '#94a3b8',
}

const AGENT_COLORS = {
  Connected: '#38bdf8',
  Silent: '#a855f7',
  'Never connected': '#94a3b8',
}

export default function DashboardPage() {
  const { isAuthenticated, isReady } = useAuth()

  const [clusters, setClusters] = useState([])
  const [agents, setAgents] = useState([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    if (!isReady || !isAuthenticated) return

    const loadStats = async () => {
      setIsLoading(true)

      try {
        const [clustersRes, agentsRes] = await Promise.all([
          apiFetch('/api/clusters'),
          apiFetch('/api/agents'),
        ])

        const clustersData = await clustersRes.json().catch(() => [])
        const agentsData = await agentsRes.json().catch(() => [])

        if (clustersRes.ok) {
          setClusters(Array.isArray(clustersData) ? clustersData : [])
        }

        if (agentsRes.ok) {
          setAgents(Array.isArray(agentsData) ? agentsData : [])
        }
      } finally {
        setIsLoading(false)
      }
    }

    loadStats()
  }, [isReady, isAuthenticated])

  const clusterSummary = useMemo(() => buildClusterSummary(clusters), [clusters])
  const agentSummary = useMemo(() => buildAgentSummary(agents), [agents])

  const clusterChartData = [
    {
      name: 'Active',
      value: clusterSummary.active,
      color: CLUSTER_COLORS['Active'],
    },
    {
      name: 'Inactive',
      value: clusterSummary.inactive,
      color: CLUSTER_COLORS['Inactive'],
    },
    {
      name: 'Never connected',
      value: clusterSummary.neverConnected,
      color: CLUSTER_COLORS['Never connected'],
    },
  ]

  const agentChartData = [
    {
      name: 'Connected',
      value: agentSummary.connected,
      color: AGENT_COLORS['Connected'],
    },
    {
      name: 'Silent',
      value: agentSummary.silent,
      color: AGENT_COLORS['Silent'],
    },
    {
      name: 'Never connected',
      value: agentSummary.neverConnected,
      color: AGENT_COLORS['Never connected'],
    },
  ]

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Overview</h1>
          <p>System summary and entry points.</p>
        </div>
      </div>

      <section className="dashboard-cards">
        <DonutSummaryCard
          icon={<FiBox />}
          title="Clusters"
          total={clusterSummary.total}
          isLoading={isLoading}
          chartData={clusterChartData}
          metaRows={[
            { label: 'Assigned agent', value: clusterSummary.assigned },
            { label: 'Unassigned', value: clusterSummary.unassigned },
          ]}
        />

        <DonutSummaryCard
          icon={<FiCpu />}
          title="Agents"
          total={agentSummary.total}
          isLoading={isLoading}
          chartData={agentChartData}
          metaRows={[
            { label: 'Enabled', value: agentSummary.enabled },
            { label: 'Disabled', value: agentSummary.disabled },
          ]}
        />
      </section>

      <section className="dashboard-cards">
        <div className="dashboard-card">
          <div className="agents-panel-header">
            <div>
              <h3>Clusters</h3>
              <p>Manage clusters, assignments, and access.</p>
            </div>

            <NavLink
              to="/dashboard/clusters"
              className="primary-button dashboard-link-button"
            >
              <FiArrowRight />
              <span>Open</span>
            </NavLink>
          </div>
        </div>

        <div className="dashboard-card">
          <div className="agents-panel-header">
            <div>
              <h3>Agents</h3>
              <p>Create and manage cluster agents.</p>
            </div>

            <NavLink
              to="/dashboard/agents"
              className="primary-button dashboard-link-button"
            >
              <FiArrowRight />
              <span>Open</span>
            </NavLink>
          </div>
        </div>
      </section>
    </>
  )
}

function DonutSummaryCard({
  icon,
  title,
  total,
  isLoading,
  chartData,
  metaRows = [],
}) {
  const visibleData = chartData.filter((item) => item.value > 0)
  const emptyData = [{ name: 'Empty', value: 1, color: '#334155' }]
  const pieData = visibleData.length > 0 ? visibleData : emptyData

  return (
    <div className="dashboard-card dashboard-card-primary donut-summary-card">
      <div className="donut-summary-header">
        <div className="donut-summary-icon">{icon}</div>

        <div>
          <h3>{title}</h3>
          <p>{isLoading ? 'Loading...' : `${total} total`}</p>
        </div>
      </div>

      <div className="donut-chart-wrap">
        <ResponsiveContainer width="100%" height={220}>
          <PieChart>
            <Pie
              data={pieData}
              dataKey="value"
              nameKey="name"
              innerRadius={52}
              outerRadius={82}
              paddingAngle={3}
              cornerRadius={6}
            >
              {pieData.map((entry) => (
                <Cell
                  key={entry.name}
                  fill={entry.color}
                />
              ))}
            </Pie>

            <Tooltip content={<ChartTooltip />} />

            <text
              x="50%"
              y="47%"
              textAnchor="middle"
              dominantBaseline="central"
              className="donut-center-total"
            >
              {isLoading ? '...' : total}
            </text>

            <text
              x="50%"
              y="57%"
              textAnchor="middle"
              dominantBaseline="central"
              className="donut-center-label"
            >
              {title}
            </text>
          </PieChart>
        </ResponsiveContainer>
      </div>

      <div className="donut-legend">
        {chartData.map((item) => (
          <div key={item.name} className="donut-legend-item">
            <span
              className="donut-legend-dot"
              style={{ backgroundColor: item.color }}
            />
            <span>{item.name}</span>
            <strong>{isLoading ? '...' : item.value}</strong>
          </div>
        ))}
      </div>

      <div className="donut-meta">
        {metaRows.map((row) => (
          <div key={row.label} className="donut-meta-row">
            <span>{row.label}</span>
            <strong>{isLoading ? '...' : row.value}</strong>
          </div>
        ))}
      </div>
    </div>
  )
}

function ChartTooltip({ active, payload }) {
  if (!active || !payload?.length) return null

  const item = payload[0]

  return (
    <div className="chart-tooltip">
      <div className="chart-tooltip-value">{item.value}</div>
      <div className="chart-tooltip-label">{item.name}</div>
    </div>
  )
}

function buildClusterSummary(clusters) {
  return clusters.reduce(
    (acc, cluster) => {
      acc.total += 1

      const hasHeartbeat = Boolean(cluster.lastHeartbeatAt)
      const assigned = Boolean(cluster.agentId)

      if (assigned) {
        acc.assigned += 1
      } else {
        acc.unassigned += 1
      }

      if (!hasHeartbeat) {
        acc.neverConnected += 1
        return acc
      }

      if (isClusterActive(cluster)) {
        acc.active += 1
      } else {
        acc.inactive += 1
      }

      return acc
    },
    {
      total: 0,
      active: 0,
      inactive: 0,
      neverConnected: 0,
      assigned: 0,
      unassigned: 0,
    }
  )
}

function buildAgentSummary(agents) {
  return agents.reduce(
    (acc, agent) => {
      acc.total += 1

      if (agent.enabled) {
        acc.enabled += 1
      } else {
        acc.disabled += 1
      }

      if (!agent.lastHeartbeatAt) {
        acc.neverConnected += 1
        return acc
      }

      if (agent.enabled && isFreshHeartbeat(agent.lastHeartbeatAt)) {
        acc.connected += 1
      } else {
        acc.silent += 1
      }

      return acc
    },
    {
      total: 0,
      connected: 0,
      silent: 0,
      neverConnected: 0,
      enabled: 0,
      disabled: 0,
    }
  )
}

function isClusterActive(cluster) {
  const normalizedStatus = String(cluster.status || '').toLowerCase()

  if (['online', 'running', 'healthy', 'active'].includes(normalizedStatus)) {
    return true
  }

  if (['offline', 'inactive', 'error', 'failed'].includes(normalizedStatus)) {
    return false
  }

  return isFreshHeartbeat(cluster.lastHeartbeatAt)
}

function isFreshHeartbeat(value) {
  if (!value) return false

  const timestamp = new Date(value).getTime()
  if (Number.isNaN(timestamp)) return false

  return Date.now() - timestamp <= HEARTBEAT_FRESH_MS
}