import { Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardLayout from './pages/DashboardLayout'
import DashboardPage from './pages/dashboard/DashboardPage'
import ClusterMetricsDetailPage from './pages/dashboard/ClusterMetricsDetailPage'
import ClusterInventoryDetailPage from './pages/dashboard/ClusterInventoryDetailPage'
import ClusterRecommendationsPage from './pages/dashboard/ClusterRecommendationsPage'
import SettingsPage from './pages/dashboard/SettingsPage'
import ProtectedRoute from './components/ProtectedRoute'
import HelpPage from './pages/dashboard/HelpPage'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />

      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />

      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <DashboardLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<DashboardPage />} />

        <Route
          path="clusters/:clusterId/metrics"
          element={<ClusterMetricsDetailPage />}
        />
        <Route
          path="clusters/:clusterId/inventory"
          element={<ClusterInventoryDetailPage />}
        />
        <Route
          path="clusters/:clusterId/recommendations"
          element={<ClusterRecommendationsPage />}
        />

        <Route path="settings" element={<SettingsPage />} />
        <Route path="help" element={<HelpPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}

export default App