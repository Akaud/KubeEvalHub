import { useAuth } from '../context/AuthContext'
import { useNavigate } from 'react-router-dom'

export default function DashboardPage() {
  const { logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div>
      <h1>Dashboard</h1>
      <p>You are authenticated</p>
      <button onClick={handleLogout}>Logout</button>
    </div>
  )
}