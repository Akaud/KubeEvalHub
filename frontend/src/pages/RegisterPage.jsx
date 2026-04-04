import { useState } from 'react'
import { FiEye, FiEyeOff } from 'react-icons/fi'
import '../styles/RegisterPage.css'
import { useNavigate } from 'react-router-dom'

export default function RegisterPage() {
  const [form, setForm] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    consent: false,
  })

  const [isEmailValid, setIsEmailValid] = useState(false)
  const [passwordsMatch, setPasswordsMatch] = useState(true)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const navigate = useNavigate()
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    let nextValue = type === 'checkbox' ? checked : value

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
    const password = form.password.trim()
    const confirmPassword = form.confirmPassword.trim()

    setError('')
    setSuccess('')

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

    if (!form.consent) {
      setError('You must agree to the policy')
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
        setError(data?.error || 'Registration failed')
        return
      }

      setSuccess('Registration successful. Redirecting to login...')

      setForm({
        username: '',
        email: '',
        password: '',
        confirmPassword: '',
        consent: false,
      })
      setIsEmailValid(false)
      setPasswordsMatch(true)

      setTimeout(() => {
        navigate('/login', { replace: true })
      }, 1500)
    } catch {
      setError('Network error')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="auth-container">
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
          {isEmailValid && <span className="check">✔</span>}
        </div>

        <div className="input-group">
          <input
            type={showPassword ? 'text' : 'password'}
            name="password"
            placeholder="Password"
            value={form.password}
            onChange={handleChange}
            required
            minLength={8}
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
            minLength={8}
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

        {error && <p className="error-text">{error}</p>}
        {success && <p className="success-text">{success}</p>}

        <label className="checkbox">
          <input
            type="checkbox"
            name="consent"
            checked={form.consent}
            onChange={handleChange}
            required
          />
          <span>
            I agree to have my data analyzed according to the policy
          </span>
        </label>

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