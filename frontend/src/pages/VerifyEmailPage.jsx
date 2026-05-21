import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import '../styles/VerifyEmailPage.css'

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const [status, setStatus] = useState('verifying')
  const [message, setMessage] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    const token = searchParams.get('token')
    
    if (!token) {
      setStatus('error')
      setMessage('Invalid verification link')
      return
    }

    const verifyEmail = async () => {
      try {
        const res = await fetch('/api/auth/verify-email', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ token }),
        })

        const data = await res.json().catch(() => null)

        if (res.ok) {
          setStatus('success')
          setMessage(data?.message || 'Email verified successfully!')
          setTimeout(() => {
            navigate('/login', { 
              state: { message: 'Email verified! You can now log in.' }
            })
          }, 3000)
        } else {
          setStatus('error')
          setMessage(data?.error || 'Invalid or expired verification link')
        }
      } catch {
        setStatus('error')
        setMessage('Network error. Please try again.')
      }
    }

    verifyEmail()
  }, [searchParams, navigate])

  return (
    <div className="verify-container">
      <div className="verify-card">
        {status === 'verifying' && (
          <>
            <div className="spinner"></div>
            <h2>Verifying your email...</h2>
          </>
        )}
        
        {status === 'success' && (
          <>
            <div className="success-icon">✓</div>
            <h2>Email Verified!</h2>
            <p>{message}</p>
            <p>Redirecting to login...</p>
          </>
        )}
        
        {status === 'error' && (
          <>
            <div className="error-icon">✗</div>
            <h2>Verification Failed</h2>
            <p>{message}</p>
            <button onClick={() => navigate('/login')}>
              Go to Login
            </button>
            <button 
              className="resend-btn"
              onClick={() => navigate('/resend-verification')}
            >
              Resend Verification Email
            </button>
          </>
        )}
      </div>
    </div>
  )
}