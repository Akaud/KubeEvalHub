import { useState } from 'react'
import '../styles/LoginPage.css'
import { useNavigate } from 'react-router-dom'

export default function LoginPage({ goToRegister }) {
  const [form, setForm] = useState({
    email: '',
    password: '',
  })

  const [showPassword, setShowPassword] = useState(false)
  const navigate = useNavigate()

  const handleChange = (e) => {
    const { name, value } = e.target
    setForm((prev) => ({ ...prev, [name]: value }))
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    console.log(form)
  }

  return (
    <div className="auth-container">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Login</h1>

        <input
          type="email"
          name="email"
          placeholder="Email"
          value={form.email}
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