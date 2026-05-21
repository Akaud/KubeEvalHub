import { useEffect, useState } from 'react'
import { FiEye, FiEyeOff } from 'react-icons/fi'
import '../styles/LoginPage.css'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

export default function LoginPage() {
  const [form, setForm] = useState({
    identifier: '',
    password: '',
  })
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [successMessage, setSuccessMessage] = useState('')
  const navigate = useNavigate()
  const location = useLocation()
  const { login, isAuthenticated, isReady } = useAuth()

  // Check for success message from navigation state (e.g., after email verification)
  useEffect(() => {
    if (location.state?.message) {
      setSuccessMessage(location.state.message)
      // Clear the state after showing
      window.history.replaceState({}, document.title)
    }
  }, [location])

  // Redirect if already authenticated
  useEffect(() => {
    if (isReady && isAuthenticated) {
      navigate('/dashboard/profile', { replace: true })
    }
  }, [isAuthenticated, isReady, navigate])

  // Clear messages after 5 seconds
  useEffect(() => {
    if (!successMessage) return

    const timer = setTimeout(() => {
      setSuccessMessage('')
    }, 5000)

    return () => clearTimeout(timer)
  }, [successMessage])

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
    setSuccessMessage('')

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ identifier, password }),
      })

      const data = await res.json().catch(() => null)

      if (!res.ok) {
        // Handle specific error cases
        if (res.status === 403 && data?.error?.includes('email not verified')) {
          setError('Email not verified. Please check your inbox for the verification link.')
          // Optionally add a button to resend verification
        } else {
          setError(data?.error || 'Login failed')
        }
        return
      }

      if (!data?.accessToken || !data?.refreshToken) {
        setError('Login response is invalid')
        return
      }

      login(data.accessToken, data.refreshToken)
      navigate('/dashboard/profile', { replace: true })
    } catch {
      setError('Network error. Please try again.')
    }
  }

  return (
    <div className="auth-container">
      {successMessage && (
        <div className="toast toast-success">
          {successMessage}
        </div>
      )}
      
      {error && (
        <div className="toast toast-error">
          {error}
          {error.includes('Email not verified') && (
            <button 
              className="resend-link"
              onClick={async () => {
                try {
                  const res = await fetch('/api/auth/resend-verification', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email: form.identifier }),
                  })
                  
                  if (res.ok) {
                    setSuccessMessage('Verification email resent! Please check your inbox.')
                    setError('')
                  } else {
                    const data = await res.json()
                    setError(data?.error || 'Failed to resend verification email')
                  }
                } catch {
                  setError('Network error. Please try again.')
                }
              }}
            >
              Resend verification email
            </button>
          )}
        </div>
      )}

      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Login</h1>

        <div className="input-group">
          <input
            type="text"
            name="identifier"
            placeholder="Email or username"
            value={form.identifier}
            onChange={handleChange}
            required
            autoComplete="username"
          />
        </div>

        <div className="input-group">
          <input
            type={showPassword ? 'text' : 'password'}
            name="password"
            placeholder="Password"
            value={form.password}
            onChange={handleChange}
            required
            autoComplete="current-password"
          />
          <span
            className="eye"
            onClick={() => setShowPassword((prev) => !prev)}
          >
            {showPassword ? <FiEyeOff /> : <FiEye />}
          </span>
        </div>

        <button type="submit">Login</button>

        <p className="switch-text link" onClick={() => navigate('/register')}>
          Don't have an account?
        </p>
      </form>
    </div>
  )
}