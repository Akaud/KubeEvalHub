import { useEffect, useState } from 'react'
import { FiSun, FiMoon } from 'react-icons/fi'

export default function SettingsPage() {
  const [theme, setTheme] = useState(localStorage.getItem('theme') || 'light')

  useEffect(() => {
    document.body.classList.remove('theme-light', 'theme-dark')
    document.body.classList.add(theme === 'dark' ? 'theme-dark' : 'theme-light')
    localStorage.setItem('theme', theme)
    window.dispatchEvent(new Event('theme-change'))
  }, [theme])

  return (
    <section className="settings-panel">
      <div className="settings-card">
        <h3>Appearance</h3>
        <p>Choose the dashboard theme.</p>

        <div className="theme-toggle">
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
      </div>
    </section>
  )
}