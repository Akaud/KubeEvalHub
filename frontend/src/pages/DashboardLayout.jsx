import { useEffect, useState } from 'react'
import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import {
  FiUser,
  FiServer,
  FiCpu,
  FiSettings,
  FiLogOut,
  FiSun,
  FiMoon,
} from 'react-icons/fi'
import '../styles/DashboardPage.css'

export default function DashboardLayout() {
  const [theme, setTheme] = useState(localStorage.getItem('theme') || 'light')
  const { logout } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    document.body.classList.remove('theme-light', 'theme-dark')
    document.body.classList.add(theme === 'dark' ? 'theme-dark' : 'theme-light')
    localStorage.setItem('theme', theme)
  }, [theme])

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className={`dashboard-layout ${theme === 'dark' ? 'dashboard-layout-dark' : ''}`}>
      <aside className="dashboard-sidebar">
        <div className="sidebar-brand">
          <h2>Donezo</h2>
          <p>Toolbox</p>
        </div>

        <nav className="sidebar-nav">
          <NavLink
            to="/dashboard/profile"
            className={({ isActive }) => `sidebar-item ${isActive ? 'active' : ''}`}
          >
            <FiUser />
            <span>Profile</span>
          </NavLink>

          <NavLink
            to="/dashboard/clusters"
            className={({ isActive }) => `sidebar-item ${isActive ? 'active' : ''}`}
          >
            <FiServer />
            <span>Clusters</span>
          </NavLink>

          <NavLink
            to="/dashboard/agents"
            className={({ isActive }) => `sidebar-item ${isActive ? 'active' : ''}`}
          >
            <FiCpu />
            <span>Agents</span>
          </NavLink>

          <NavLink
            to="/dashboard/settings"
            className={({ isActive }) => `sidebar-item ${isActive ? 'active' : ''}`}
          >
            <FiSettings />
            <span>Settings</span>
          </NavLink>
        </nav>

        <div className="sidebar-theme">
          <button
            type="button"
            className={`theme-button ${theme === 'light' ? 'active' : ''}`}
            onClick={() => setTheme('light')}
          >
            <FiSun />
            <span>Light</span>
          </button>

          <button
            type="button"
            className={`theme-button ${theme === 'dark' ? 'active' : ''}`}
            onClick={() => setTheme('dark')}
          >
            <FiMoon />
            <span>Dark</span>
          </button>
        </div>

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