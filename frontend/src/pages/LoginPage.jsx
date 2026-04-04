import { useState } from 'react'
import '../styles/LoginPage.css'
import { useNavigate } from 'react-router-dom'

export default function LoginPage({ goToRegister }) {
  const [form, setForm] = useState({
    identifier: '',
    password: '',
  })

  const [showPassword, setShowPassword] = useState(false)
  const navigate = useNavigate()

  const handleChange = (e) => {
    const { name, value } = e.target

    let nextValue = value
    if (name === 'identifier') {
      nextValue = nextValue.trimStart()
    }

    setForm((prev) => ({ ...prev, [name]: nextValue }))
  }

  const handleSubmit = (e) => {
    e.preventDefault()

    const identifier = form.identifier.trim()
    const password = form.password.trim()

    if (!identifier || !password) return

    console.log({ identifier, password })
  }

  return (
    <div className="auth-container">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Login</h1>

        <input
          type="text"
          name="identifier"
          placeholder="Email or Username"
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

        <button type="submit">Login</button>

        <p className="switch-text link" onClick={() => navigate('/register')}>
          Don't have an account?
        </p>
      </form>
    </div>
  )
}