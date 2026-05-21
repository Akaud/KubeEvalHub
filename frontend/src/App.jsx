import { Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import VerifyEmailPage from './pages/VerifyEmailPage'
import ResendVerificationPage from './pages/ResendVerificationPage'
import DashboardLayout from './pages/DashboardLayout'
import DashboardPage from './pages/dashboard/DashboardPage'
import ClustersPage from './pages/dashboard/ClustersPage'
import AgentsPage from './pages/dashboard/AgentsPage'
import ClusterMetricsDetailPage from './pages/dashboard/ClusterMetricsDetailPage'
import ClusterInventoryDetailPage from './pages/dashboard/ClusterInventoryDetailPage'
import ClusterRecommendationsPage from './pages/dashboard/ClusterRecommendationsPage'
import SettingsPage from './pages/dashboard/SettingsPage'
import HelpPage from './pages/dashboard/HelpPage'
import GuidePage from './pages/dashboard/GuidePage'
import ProtectedRoute from './components/ProtectedRoute'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />

      {/* Public routes */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/verify-email" element={<VerifyEmailPage />} />
      <Route path="/resend-verification" element={<ResendVerificationPage />} />

      {/* Protected routes */}
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <DashboardLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="clusters" element={<ClustersPage />} />
        <Route path="agents" element={<AgentsPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="guide" element={<GuidePage />} />
        <Route path="help" element={<HelpPage />} />

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
      </Route>

      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}

export default App