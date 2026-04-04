import { useEffect, useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { useNavigate } from 'react-router-dom'
import { FiUser, FiServer, FiSettings, FiLogOut, FiSun, FiMoon } from 'react-icons/fi'
import '../styles/DashboardPage.css'

export default function DashboardPage() {
  const [activeSection, setActiveSection] = useState('profile')
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

  const sectionContent = {
    profile: {
      title: 'Profile',
      description: 'View and manage your account information.',
      cards: [
        {
          title: 'Profile',
          text: 'View and manage your account information.',
          primary: true,
        },
        {
          title: 'Account',
          text: 'Inspect personal details and authentication data.',
        },
        {
          title: 'Security',
          text: 'Review password and access-related settings.',
        },
      ],
    },
    cluster: {
      title: 'Cluster',
      description: 'Inspect cluster state, services, and workloads.',
      cards: [
        {
          title: 'Nodes',
          text: 'Review node health, capacity, and availability.',
          primary: true,
        },
        {
          title: 'Workloads',
          text: 'Track deployments, jobs, and running services.',
        },
        {
          title: 'Monitoring',
          text: 'Inspect logs, metrics, and runtime signals.',
        },
      ],
    },
    settings: {
      title: 'Settings',
      description: 'Update system preferences and configuration.',
      cards: [],
    },
  }

  const currentSection = sectionContent[activeSection]

  return (
    <div className={`dashboard-layout ${theme === 'dark' ? 'dashboard-layout-dark' : ''}`}>
      <aside className="dashboard-sidebar">
        <div className="sidebar-brand">
          <h2>Donezo</h2>
          <p>Toolbox</p>
        </div>

        <nav className="sidebar-nav">
          <button
            className={`sidebar-item ${activeSection === 'profile' ? 'active' : ''}`}
            type="button"
            onClick={() => setActiveSection('profile')}
          >
            <FiUser />
            <span>Profile</span>
          </button>

          <button
            className={`sidebar-item ${activeSection === 'cluster' ? 'active' : ''}`}
            type="button"
            onClick={() => setActiveSection('cluster')}
          >
            <FiServer />
            <span>Cluster</span>
          </button>

          <button
            className={`sidebar-item ${activeSection === 'settings' ? 'active' : ''}`}
            type="button"
            onClick={() => setActiveSection('settings')}
          >
            <FiSettings />
            <span>Settings</span>
          </button>
        </nav>

        <button className="sidebar-logout" type="button" onClick={handleLogout}>
          <FiLogOut />
          <span>Logout</span>
        </button>
      </aside>

      <main className="dashboard-content">
        <div className="dashboard-header">
          <div>
            <h1>{currentSection.title}</h1>
            <p>{currentSection.description}</p>
          </div>
        </div>

        {activeSection !== 'settings' ? (
          <section className="dashboard-cards">
            {currentSection.cards.map((card) => (
              <div
                key={card.title}
                className={`dashboard-card ${card.primary ? 'dashboard-card-primary' : ''}`}
              >
                <h3>{card.title}</h3>
                <p>{card.text}</p>
              </div>
            ))}
          </section>
        ) : (
          <section className="settings-panel">
            <div className="settings-card">
              <h3>Appearance</h3>
              <p>Choose how the dashboard should look.</p>

              <div className="theme-toggle">
                <button
                  type="button"
                  className={`theme-button ${theme === 'light' ? 'active' : ''}`}
                  onClick={() => setTheme('light')}
                >
                  <FiSun />
                  <span>Light mode</span>
                </button>

                <button
                  type="button"
                  className={`theme-button ${theme === 'dark' ? 'active' : ''}`}
                  onClick={() => setTheme('dark')}
                >
                  <FiMoon />
                  <span>Dark mode</span>
                </button>
              </div>
            </div>
          </section>
        )}
      </main>
    </div>
  )
}