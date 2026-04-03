import { useState } from 'react'
import { FiEye, FiEyeOff } from 'react-icons/fi'
import '../styles/RegisterPage.css'
import { useNavigate } from 'react-router-dom'

export default function RegisterPage({ goToLogin }) {
  const [form, setForm] = useState({
    email: '',
    password: '',
    confirmPassword: '',
    consent: false,
  })

  const [isEmailValid, setIsEmailValid] = useState(false)
  const [passwordsMatch, setPasswordsMatch] = useState(true)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const navigate = useNavigate()

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    let nextValue = type === 'checkbox' ? checked : value

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
  }

  const handleSubmit = (e) => {
    e.preventDefault()

    const email = form.email.trim()
    const password = form.password.trim()
    const confirmPassword = form.confirmPassword.trim()

    if (!email || !password || !confirmPassword) return
    if (!emailRegex.test(email)) return
    if (password !== confirmPassword) return
    if (!form.consent) return

    console.log({ email, password })
  }

  return (
    <div className="auth-container">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Register</h1>

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

        <button type="submit">Register</button>

        <p className="switch-text link" onClick={() => navigate('/login')}>
          I already have an account
        </p>
      </form>
    </div>
  )
}