import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import '../styles/LoginPage.css'

export default function ResendVerificationPage() {
  const [email, setEmail] = useState('')
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e) => {
    e.preventDefault()
    setMessage('')
    setError('')
    setIsSubmitting(true)

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(email)) {
      setError('Please enter a valid email address')
      setIsSubmitting(false)
      return
    }

    try {
      const res = await fetch('/api/auth/resend-verification', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      })

      const data = await res.json().catch(() => null)

      if (res.ok) {
        setMessage(data?.message || 'Verification email sent! Please check your inbox.')
        setEmail('')
        setTimeout(() => {
          navigate('/login', { 
            state: { message: 'Verification email sent! Please check your inbox.' }
          })
        }, 3000)
      } else if (res.status === 404) {
        setError('No account found with this email address.')
      } else if (res.status === 409) {
        setError('This email is already verified. Please login.')
        setTimeout(() => {
          navigate('/login', { 
            state: { message: 'This email is already verified. You can login now.' }
          })
        }, 2000)
      } else {
        setError(data?.error || 'Failed to send verification email. Please try again.')
      }
    } catch {
      setError('Network error. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="auth-container">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Resend Verification Email</h1>
        
        <p style={{ textAlign: 'center', marginBottom: '20px', color: '#666' }}>
          Enter your email to receive a new verification link
        </p>
        
        {message && <div className="toast toast-success">{message}</div>}
        {error && <div className="toast toast-error">{error}</div>}

        <div className="input-group">
          <input
            type="email"
            placeholder="Email address"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            disabled={isSubmitting}
          />
        </div>
        
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Sending...' : 'Resend Verification Email'}
        </button>
        
        <p className="switch-text link" onClick={() => navigate('/login')}>
          Back to Login
        </p>
      </form>
    </div>
  )
}