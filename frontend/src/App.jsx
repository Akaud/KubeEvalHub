import { Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardLayout from './pages/DashboardLayout'
import ProfilePage from './pages/dashboard/ProfilePage'
import ClustersPage from './pages/dashboard/ClustersPage'
import ClusterMetricsDetailPage from './pages/dashboard/ClusterMetricsDetailPage'
import AgentsPage from './pages/dashboard/AgentsPage'
import SettingsPage from './pages/dashboard/SettingsPage'
import ProtectedRoute from './components/ProtectedRoute'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/login" replace />} />

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
        <Route index element={<Navigate to="/dashboard/profile" replace />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route path="clusters" element={<ClustersPage />} />
        <Route path="clusters/:agentId/metrics" element={<ClusterMetricsDetailPage />} />
        <Route path="agents" element={<AgentsPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/dashboard/profile" replace />} />
    </Routes>
  )
}

export default App