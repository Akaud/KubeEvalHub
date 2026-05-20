import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { FiAlertTriangle } from 'react-icons/fi'
import { useAuth } from '../../context/AuthContext'
import { apiFetch } from '../../utils/apiFetch'
import { applyTheme, getStoredTheme } from '../../utils/theme'

export default function SettingsPage() {
  const { isAuthenticated, isReady, logout } = useAuth()
  const navigate = useNavigate()

  const [theme, setTheme] = useState(getStoredTheme())

  useEffect(() => {
    applyTheme(theme)
  }, [theme])

  const [currentUser, setCurrentUser] = useState(null)
  const [isLoadingUser, setIsLoadingUser] = useState(false)

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const [isSavingProfile, setIsSavingProfile] = useState(false)
  const [isSavingPassword, setIsSavingPassword] = useState(false)
  const [isDeletingAccount, setIsDeletingAccount] = useState(false)

  const [profileError, setProfileError] = useState('')
  const [passwordError, setPasswordError] = useState('')
  const [deleteError, setDeleteError] = useState('')
  const [profileSuccess, setProfileSuccess] = useState('')
  const [passwordSuccess, setPasswordSuccess] = useState('')

  const [showPasswordConfirm, setShowPasswordConfirm] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const notifyProfileUpdated = useCallback(() => {
    window.dispatchEvent(new Event('user-profile-updated'))
  }, [])

  const loadCurrentUser = useCallback(async () => {
    if (!isAuthenticated) return

    setIsLoadingUser(true)
    setProfileError('')

    try {
      const res = await apiFetch('/api/users/me')
      const data = await res.json().catch(() => null)

      if (!res.ok) {
        throw new Error(data?.error || 'Failed to load current user')
      }

      setCurrentUser(data)
    } catch (err) {
      setProfileError(err.message || 'Failed to load current user')
      setCurrentUser(null)
    } finally {
      setIsLoadingUser(false)
    }
  }, [isAuthenticated])

  useEffect(() => {
    if (!isReady || !isAuthenticated) return
    loadCurrentUser()
  }, [isReady, isAuthenticated, loadCurrentUser])

  async function patchUser(fields) {
    if (!isAuthenticated) {
      throw new Error('Missing authentication')
    }

    if (!currentUser?.id) {
      throw new Error('Missing current user id')
    }

    const res = await apiFetch(`/api/users/${currentUser.id}`, {
      method: 'PATCH',
      body: JSON.stringify(fields),
    })

    if (res.status === 204) {
      return null
    }

    const data = await res.json().catch(() => null)

    if (!res.ok) {
      throw new Error(data?.error || 'Failed to update profile')
    }

    return data
  }

  async function handleProfileSubmit(e) {
    e.preventDefault()
    setProfileError('')
    setProfileSuccess('')

    const trimmedName = name.trim()
    const trimmedEmail = email.trim()

    const body = {}

    if (trimmedName && trimmedName !== currentUser?.name) {
      body.name = trimmedName
    }

    if (trimmedEmail && trimmedEmail !== currentUser?.email) {
      body.email = trimmedEmail
    }

    if (!body.name && !body.email) {
      setProfileError('No profile changes to save')
      return
    }

    setIsSavingProfile(true)

    try {
      const updatedUser = await patchUser(body)

      if (updatedUser) {
        setCurrentUser(updatedUser)
      } else {
        await loadCurrentUser()
      }

      notifyProfileUpdated()
      setProfileSuccess('Profile updated successfully')
      setName('')
      setEmail('')
    } catch (err) {
      setProfileError(err.message || 'Failed to update profile')
    } finally {
      setIsSavingProfile(false)
    }
  }

  function handlePasswordSubmit(e) {
    e.preventDefault()
    setPasswordError('')
    setPasswordSuccess('')

    if (!password) {
      setPasswordError('Enter a new password')
      return
    }

    setShowPasswordConfirm(true)
  }

  async function confirmPasswordChange() {
    setPasswordError('')
    setPasswordSuccess('')
    setIsSavingPassword(true)

    try {
      await patchUser({ password })
      setPasswordSuccess('Password updated successfully')
      setPassword('')
      setShowPasswordConfirm(false)
    } catch (err) {
      setPasswordError(err.message || 'Failed to update password')
    } finally {
      setIsSavingPassword(false)
    }
  }

  function cancelPasswordChange() {
    if (isSavingPassword) return
    setShowPasswordConfirm(false)
  }

  function handleDeleteAccountClick() {
    setDeleteError('')
    setShowDeleteConfirm(true)
  }

  function cancelDeleteAccount() {
    if (isDeletingAccount) return
    setShowDeleteConfirm(false)
  }

  async function confirmDeleteAccount() {
    setDeleteError('')
    setIsDeletingAccount(true)

    try {
      const res = await apiFetch('/api/users/me', {
        method: 'DELETE',
      })

      const data = await res.json().catch(() => null)

      if (!res.ok) {
        throw new Error(data?.error || 'Failed to delete account')
      }

      await logout()
      navigate('/login', { replace: true })
    } catch (err) {
      setDeleteError(err.message || 'Failed to delete account')
      setShowDeleteConfirm(false)
    } finally {
      setIsDeletingAccount(false)
    }
  }

  return (
    <>
      <div className="dashboard-header">
        <div>
          <h1>Account settings</h1>
          <p>Manage account information and security preferences.</p>
        </div>
      </div>

      <section className="settings-panel settings-grid">
        <div className="settings-card">
          <h3>Account</h3>
          <p>Change your username and email.</p>

          <form className="profile-form" onSubmit={handleProfileSubmit}>
            <div className="field-group">
              <label htmlFor="settings-name">New username</label>
              <input
                id="settings-name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={currentUser?.name || 'Enter new username'}
                disabled={isSavingProfile || isLoadingUser || !currentUser?.id}
              />
            </div>

            <div className="field-group">
              <label htmlFor="settings-email">New email</label>
              <input
                id="settings-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder={currentUser?.email || 'Enter new email'}
                disabled={isSavingProfile || isLoadingUser || !currentUser?.id}
              />
            </div>

            {profileError && <div className="form-error">{profileError}</div>}
            {profileSuccess && <div className="success-box">{profileSuccess}</div>}

            <div className="profile-actions">
              <button
                type="submit"
                className="primary-button"
                disabled={isSavingProfile || isLoadingUser || !currentUser?.id}
              >
                {isSavingProfile ? 'Saving...' : 'Save account changes'}
              </button>
            </div>
          </form>
        </div>

        <div className="settings-card">
          <h3>Security</h3>
          <p>Change your password.</p>

          <form className="profile-form" onSubmit={handlePasswordSubmit}>
            <div className="field-group">
              <label htmlFor="settings-password">New password</label>
              <input
                id="settings-password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter new password"
                disabled={isSavingPassword || isLoadingUser || !currentUser?.id}
              />
            </div>

            {passwordError && <div className="form-error">{passwordError}</div>}
            {passwordSuccess && <div className="success-box">{passwordSuccess}</div>}

            <div className="profile-actions">
              <button
                type="submit"
                className="primary-button"
                disabled={isSavingPassword || isLoadingUser || !currentUser?.id}
              >
                Change password
              </button>
            </div>
          </form>
        </div>

        <div className="settings-card danger-card">
          <h3>Delete account</h3>
          <p>Permanently delete your account. This action cannot be undone.</p>

          {deleteError && <div className="form-error">{deleteError}</div>}

          <div className="profile-actions">
            <button
              type="button"
              className="danger-button"
              onClick={handleDeleteAccountClick}
              disabled={isDeletingAccount || isLoadingUser || !currentUser?.id}
            >
              <FiAlertTriangle />
              <span>Delete my account</span>
            </button>
          </div>
        </div>
      </section>

      {showPasswordConfirm && (
        <div className="modal-backdrop">
          <div className="modal-card profile-confirm-modal">
            <div className="modal-header">
              <div>
                <h3>Confirm password change</h3>
                <p>
                  Are you sure you want to change your password? This action will update
                  your account credentials.
                </p>
              </div>
            </div>

            <div className="modal-actions">
              <button
                type="button"
                className="secondary-button"
                onClick={cancelPasswordChange}
                disabled={isSavingPassword}
              >
                Cancel
              </button>

              <button
                type="button"
                className="primary-button"
                onClick={confirmPasswordChange}
                disabled={isSavingPassword}
              >
                {isSavingPassword ? 'Saving...' : 'OK'}
              </button>
            </div>
          </div>
        </div>
      )}

      {showDeleteConfirm && (
        <div className="modal-backdrop">
          <div className="modal-card profile-confirm-modal">
            <div className="modal-header">
              <div>
                <h3>Confirm account deletion</h3>
                <p>
                  Are you sure you want to delete your account? This will permanently
                  remove your user account and log you out.
                </p>
              </div>
            </div>

            <div className="modal-actions">
              <button
                type="button"
                className="secondary-button"
                onClick={cancelDeleteAccount}
                disabled={isDeletingAccount}
              >
                Cancel
              </button>

              <button
                type="button"
                className="danger-button"
                onClick={confirmDeleteAccount}
                disabled={isDeletingAccount}
              >
                <FiAlertTriangle />
                <span>{isDeletingAccount ? 'Deleting...' : 'Delete account'}</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}