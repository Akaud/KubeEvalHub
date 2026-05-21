import { useEffect, useState } from 'react'
import { FiEye, FiEyeOff } from 'react-icons/fi'
import '../styles/RegisterPage.css'
import { useNavigate } from 'react-router-dom'

export default function RegisterPage() {
  const [form, setForm] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  })

  const [isEmailValid, setIsEmailValid] = useState(false)
  const [passwordsMatch, setPasswordsMatch] = useState(true)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [error, setError] = useState('')
  const [successMessage, setSuccessMessage] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const navigate = useNavigate()
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

  useEffect(() => {
    if (!successMessage) return

    const timer = setTimeout(() => {
      setSuccessMessage('')
      // Redirect to login page with a message
      navigate('/login', { 
        state: { message: 'Registration successful! Please check your email to verify your account.' }
      })
    }, 3000)

    return () => clearTimeout(timer)
  }, [successMessage, navigate])

  const handleChange = (e) => {
    const { name, value } = e.target
    let nextValue = value

    if (name === 'username') {
      nextValue = nextValue.replace(/\s/g, '')
    }

    if (name === 'email') {
      nextValue = nextValue.replace(/\s/g, '')
      setIsEmailValid(emailRegex.test(nextValue))
    }

    const updatedForm = { ...form, [name]: nextValue }
    setForm(updatedForm)

    if (
      updatedForm.password &&
      updatedForm.confirmPassword &&
      updatedForm.password !== updatedForm.confirmPassword
    ) {
      setPasswordsMatch(false)
    } else {
      setPasswordsMatch(true)
    }

    setError('')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()

    const username = form.username.trim()
    const email = form.email.trim()
    const password = form.password
    const confirmPassword = form.confirmPassword

    setError('')
    setSuccessMessage('')

    if (!username || !email || !password || !confirmPassword) {
      setError('All fields are required')
      return
    }

    if (!emailRegex.test(email)) {
      setError('Invalid email format')
      return
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    if (password.length < 12) {
      setError('Password must be at least 12 characters')
      return
    }

    try {
      setIsSubmitting(true)

      const res = await fetch('/api/users/', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: username,
          email,
          password,
        }),
      })

      const data = await res.json().catch(() => null)

      if (!res.ok) {
        if (res.status === 409) {
          setError('Email already exists. Please use a different email or login.')
        } else if (data?.error) {
          setError(data.error)
        } else {
          setError('Registration failed')
        }
        return
      }

      // Registration successful - show success message
      setSuccessMessage(data?.message || 'Registration successful! Redirecting to login...')
      
    } catch {
      setError('Network error. Please try again.')
    } finally {
      setIsSubmitting(false)
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
        </div>
      )}

      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Register</h1>

        <div className="input-group">
          <input
            type="text"
            name="username"
            placeholder="Username"
            value={form.username}
            onChange={handleChange}
            required
            minLength={3}
            maxLength={30}
          />
        </div>

        <div className="input-group">
          <input
            type="email"
            name="email"
            placeholder="Email"
            value={form.email}
            onChange={handleChange}
            required
            maxLength={254}
          />
          {isEmailValid && form.email && <span className="check">✔</span>}
        </div>

        <div className="input-group">
          <input
            type={showPassword ? 'text' : 'password'}
            name="password"
            placeholder="Password (min. 12 characters)"
            value={form.password}
            onChange={handleChange}
            required
            minLength={12}
          />
          <span
            className="eye"
            onClick={() => setShowPassword((prev) => !prev)}
          >
            {showPassword ? <FiEyeOff /> : <FiEye />}
          </span>
        </div>

        <div className="input-group password-group">
          <input
            type={showConfirmPassword ? 'text' : 'password'}
            name="confirmPassword"
            placeholder="Repeat password"
            value={form.confirmPassword}
            onChange={handleChange}
            required
            minLength={12}
          />
          <span
            className="eye"
            onClick={() => setShowConfirmPassword((prev) => !prev)}
          >
            {showConfirmPassword ? <FiEyeOff /> : <FiEye />}
          </span>
        </div>

        {!passwordsMatch && (
          <p className="error-text">Passwords do not match</p>
        )}

        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Registering...' : 'Register'}
        </button>

        <p className="switch-text link" onClick={() => navigate('/login')}>
          I already have an account
        </p>
      </form>
    </div>
  )
}