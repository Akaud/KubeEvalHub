import { useEffect, useState, useCallback, useRef, useMemo } from 'react'
import { Outlet, NavLink, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { apiFetch } from '../utils/apiFetch'
import { applyTheme, getStoredTheme } from '../utils/theme'
import {
  FiHome,
  FiSettings,
  FiHelpCircle,
  FiLogOut,
  FiSearch,
  FiUser,
  FiChevronDown,
  FiSun,
  FiMoon,
  FiBox,
  FiCpu,
  FiBookOpen,
} from 'react-icons/fi'
import {
  dashboardSearchItems,
  searchDashboardItems,
} from '../utils/dashboardSearch'
import '../styles/DashboardPage.css'

export default function DashboardLayout() {
  const { logout, isAuthenticated, isReady } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const dropdownRef = useRef(null)
  const searchRef = useRef(null)

  const [theme, setTheme] = useState(getStoredTheme())
  const [currentUser, setCurrentUser] = useState(null)
  const [isLoadingUser, setIsLoadingUser] = useState(false)
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)

  const [searchQuery, setSearchQuery] = useState('')
  const [isSearchOpen, setIsSearchOpen] = useState(false)

  const searchResults = useMemo(() => {
    return searchDashboardItems(searchQuery, dashboardSearchItems)
  }, [searchQuery])

  const loadCurrentUser = useCallback(async () => {
    if (!isAuthenticated) return

    setIsLoadingUser(true)

    try {
      const res = await apiFetch('/api/users/me')
      const data = await res.json().catch(() => null)

      if (!res.ok) {
        throw new Error(data?.error || 'Failed to load current user')
      }

      setCurrentUser(data)
    } catch (error) {
      console.error('DashboardLayout failed to load current user:', error)
      setCurrentUser(null)
    } finally {
      setIsLoadingUser(false)
    }
  }, [isAuthenticated])

  useEffect(() => {
    const handleThemeChange = () => {
      setTheme(getStoredTheme())
    }

    window.addEventListener('theme-change', handleThemeChange)
    window.addEventListener('storage', handleThemeChange)

    return () => {
      window.removeEventListener('theme-change', handleThemeChange)
      window.removeEventListener('storage', handleThemeChange)
    }
  }, [])

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setIsUserMenuOpen(false)
      }

      if (searchRef.current && !searchRef.current.contains(event.target)) {
        setIsSearchOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  useEffect(() => {
    const handleKeyDown = (event) => {
      const isCmdF =
        (event.metaKey || event.ctrlKey) &&
        event.key.toLowerCase() === 'f'

      if (isCmdF) {
        event.preventDefault()
        const input = document.getElementById('dashboard-search-input')
        input?.focus()
        setIsSearchOpen(true)
      }

      if (event.key === 'Escape') {
        setIsSearchOpen(false)
        setIsUserMenuOpen(false)
      }
    }

    document.addEventListener('keydown', handleKeyDown)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [])

  useEffect(() => {
    if (!isReady || !isAuthenticated) return
    loadCurrentUser()
  }, [isReady, isAuthenticated, location.pathname, loadCurrentUser])

  useEffect(() => {
    const handleUserProfileUpdated = () => {
      loadCurrentUser()
    }

    window.addEventListener('user-profile-updated', handleUserProfileUpdated)

    return () => {
      window.removeEventListener('user-profile-updated', handleUserProfileUpdated)
    }
  }, [loadCurrentUser])

  useEffect(() => {
    setIsUserMenuOpen(false)
    setIsSearchOpen(false)
  }, [location.pathname])

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  const handleThemeSelect = (nextTheme) => {
    applyTheme(nextTheme)
    setTheme(nextTheme)
  }

  const handleOpenSettings = () => {
    setIsUserMenuOpen(false)
    navigate('/dashboard/settings')
  }

  const handleSearchSelect = (route) => {
    setSearchQuery('')
    setIsSearchOpen(false)
    navigate(route)
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
              to="/dashboard"
              end
              className={({ isActive }) =>
                `sidebar-item ${isActive ? 'active' : ''}`
              }
            >
              <FiHome />
              <span>Overview</span>
            </NavLink>

            <NavLink
              to="/dashboard/clusters"
              className={({ isActive }) =>
                `sidebar-item ${isActive ? 'active' : ''}`
              }
            >
              <FiBox />
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
              to="/dashboard/guide"
              className={({ isActive }) =>
                `sidebar-item ${isActive ? 'active' : ''}`
              }
            >
              <FiBookOpen />
              <span>Guide</span>
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
        <header className="dashboard-topbar">
          <div className="dashboard-search-wrapper" ref={searchRef}>
            <div className="dashboard-searchbar">
              <FiSearch className="dashboard-search-icon" />
              <input
                id="dashboard-search-input"
                type="text"
                placeholder="Search pages, settings, help..."
                aria-label="Search"
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value)
                  setIsSearchOpen(true)
                }}
                onFocus={() => setIsSearchOpen(true)}
              />
            </div>

            {isSearchOpen && searchQuery.trim() && (
              <div className="dashboard-search-results">
                {searchResults.length > 0 ? (
                  searchResults.map((result) => (
                    <button
                      key={result.id}
                      type="button"
                      className="dashboard-search-result-item"
                      onClick={() => handleSearchSelect(result.route)}
                    >
                      <div className="dashboard-search-result-title">
                        {result.title}
                      </div>
                      <div className="dashboard-search-result-description">
                        {result.description}
                      </div>
                    </button>
                  ))
                ) : (
                  <div className="dashboard-search-no-results">No results</div>
                )}
              </div>
            )}
          </div>

          <div className="dashboard-topbar-actions">
            <div className="dashboard-user-dropdown" ref={dropdownRef}>
              <button
                type="button"
                className="dashboard-user-card dashboard-user-trigger"
                onClick={() => setIsUserMenuOpen((prev) => !prev)}
                aria-haspopup="menu"
                aria-expanded={isUserMenuOpen}
              >
                <div className="dashboard-user-avatar">
                  <FiUser />
                </div>

                <div className="dashboard-user-meta">
                  <strong>
                    {isLoadingUser ? 'Loading...' : currentUser?.name || 'User'}
                  </strong>
                  <span>
                    {isLoadingUser
                      ? 'Loading...'
                      : currentUser?.email || 'No email'}
                  </span>
                </div>

                <FiChevronDown
                  className={`dashboard-user-chevron ${
                    isUserMenuOpen ? 'open' : ''
                  }`}
                />
              </button>

              {isUserMenuOpen && (
                <div className="dashboard-user-menu" role="menu">
                  <button
                    type="button"
                    className="dashboard-user-menu-item"
                    onClick={handleOpenSettings}
                  >
                    <FiSettings />
                    <span>Account settings</span>
                  </button>

                  <div className="dashboard-user-menu-divider" />

                  <div className="dashboard-user-menu-section">
                    <p className="dashboard-user-menu-label">Theme</p>

                    <button
                      type="button"
                      className={`dashboard-theme-option ${
                        theme === 'light' ? 'active' : ''
                      }`}
                      onClick={() => handleThemeSelect('light')}
                    >
                      <FiSun />
                      <span>Light theme</span>
                    </button>

                    <button
                      type="button"
                      className={`dashboard-theme-option ${
                        theme === 'dark' ? 'active' : ''
                      }`}
                      onClick={() => handleThemeSelect('dark')}
                    >
                      <FiMoon />
                      <span>Dark theme</span>
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </header>

        <Outlet />
      </main>
    </div>
  )
}