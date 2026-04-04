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
  const { login, isAuthenticated } = useAuth()

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true })
    }
  }, [isAuthenticated, navigate])

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

    const email = form.identifier.trim()
    const password = form.password.trim()

    if (!email || !password) return

    setError('')

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })

      const data = await res.json()

      if (!res.ok) {
        setError(data.error || 'Login failed')
        return
      }

      login(data.token)
      navigate('/dashboard', { replace: true })
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
          placeholder="Email"
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

        {error && <p>{error}</p>}

        <button type="submit">Login</button>

        <p className="switch-text link" onClick={() => navigate('/register')}>
          Don't have an account?
        </p>
      </form>
    </div>
  )
}