import { useEffect, useState } from 'react'
import '../styles/LoginPage.css'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

export default function LoginPage() {
  const [form, setForm] = useState({
    identifier: '',
    password: '',
  })
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const { login, isAuthenticated, isReady } = useAuth()

  useEffect(() => {
    if (isReady && isAuthenticated) {
      navigate('/dashboard/profile', { replace: true })
    }
  }, [isAuthenticated, isReady, navigate])

  const handleChange = (e) => {
    const { name, value } = e.target

    let nextValue = value
    if (name === 'identifier') {
      nextValue = nextValue.trimStart()
    }

    setForm((prev) => ({ ...prev, [name]: nextValue }))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()

    const identifier = form.identifier.trim()
    const password = form.password

    if (!identifier || !password) return

    setError('')

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ identifier, password }),
      })

      const data = await res.json().catch(() => null)

      if (!res.ok) {
        setError(data?.error || 'Login failed')
        return
      }

      if (!data?.accessToken || !data?.refreshToken) {
        setError('Login response is invalid')
        return
      }

      login(data.accessToken, data.refreshToken)
      navigate('/dashboard/profile', { replace: true })
    } catch {
      setError('Network error')
    }
  }

  return (
    <div className="auth-container">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Login</h1>

        <input
          type="text"
          name="identifier"
          placeholder="Email or username"
          value={form.identifier}
          onChange={handleChange}
          required
        />

        <input
          type={showPassword ? 'text' : 'password'}
          name="password"
          placeholder="Password"
          value={form.password}
          onChange={handleChange}
          required
        />

        {error && <p className="error-text">{error}</p>}

        <button type="submit">Login</button>

        <p className="switch-text link" onClick={() => navigate('/register')}>
          Don't have an account?
        </p>
      </form>
    </div>
  )
}