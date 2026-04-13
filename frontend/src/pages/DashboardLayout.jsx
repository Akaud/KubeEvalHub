import { useEffect, useState } from 'react'
import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import {
  FiHome,
  FiServer,
  FiCpu,
  FiSettings,
  FiHelpCircle,
  FiLogOut,
} from 'react-icons/fi'
import '../styles/DashboardPage.css'

export default function DashboardLayout() {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const [theme, setTheme] = useState(localStorage.getItem('theme') || 'light')

  useEffect(() => {
    const handleThemeChange = () => {
      setTheme(localStorage.getItem('theme') || 'light')
    }

    window.addEventListener('theme-change', handleThemeChange)
    window.addEventListener('storage', handleThemeChange)

    return () => {
      window.removeEventListener('theme-change', handleThemeChange)
      window.removeEventListener('storage', handleThemeChange)
    }
  }, [])

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div
      className={`dashboard-layout ${
        theme === 'dark' ? 'dashboard-layout-dark' : ''
      }`}
    >
      <aside className="dashboard-sidebar">
        <div className="sidebar-brand">
          <h2>KubeEvalHub</h2>
        </div>

        <nav className="sidebar-nav">
          <div className="sidebar-section">
            <p className="sidebar-section-title">Menu</p>

            <NavLink
              to="/dashboard/clusters"
              className={({ isActive }) =>
                `sidebar-item ${isActive ? 'active' : ''}`
              }
            >
              <FiHome />
              <span>Dashboard</span>
            </NavLink>

            <NavLink
              to="/dashboard/clusters"
              className={({ isActive }) =>
                `sidebar-item ${isActive ? 'active' : ''}`
              }
            >
              <FiServer />
              <span>Clusters</span>
            </NavLink>

            <NavLink
              to="/dashboard/agents"
              className={({ isActive }) =>
                `sidebar-item ${isActive ? 'active' : ''}`
              }
            >
              <FiCpu />
              <span>Agents</span>
            </NavLink>
          </div>

          <div className="sidebar-section">
            <p className="sidebar-section-title">Tools</p>

            <NavLink
              to="/dashboard/settings"
              className={({ isActive }) =>
                `sidebar-item ${isActive ? 'active' : ''}`
              }
            >
              <FiSettings />
              <span>Settings</span>
            </NavLink>

            <NavLink
              to="/dashboard/help"
              className={({ isActive }) =>
                `sidebar-item ${isActive ? 'active' : ''}`
              }
            >
              <FiHelpCircle />
              <span>Help</span>
            </NavLink>
          </div>
        </nav>

        <button className="sidebar-logout" type="button" onClick={handleLogout}>
          <FiLogOut />
          <span>Logout</span>
        </button>
      </aside>

      <main className="dashboard-content">
        <Outlet />
      </main>
    </div>
  )
}