export async function apiFetch(url, options = {}) {
  const accessToken = localStorage.getItem('accessToken')
  const refreshToken = localStorage.getItem('refreshToken')

  const makeRequest = async (token) => {
    const headers = {
      ...(options.headers || {}),
    }

    const isFormData = options.body instanceof FormData

    if (!isFormData && !headers['Content-Type']) {
      headers['Content-Type'] = 'application/json'
    }

    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    return fetch(url, {
      ...options,
      headers,
    })
  }

  let response = await makeRequest(accessToken)

  if (response.status !== 401 || !refreshToken) {
    return response
  }

  const refreshResponse = await fetch('/api/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  })

  const refreshData = await refreshResponse.json().catch(() => null)

  if (
    !refreshResponse.ok ||
    !refreshData?.accessToken ||
    !refreshData?.refreshToken
  ) {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
    window.location.href = '/login'
    return response
  }

  localStorage.setItem('accessToken', refreshData.accessToken)
  localStorage.setItem('refreshToken', refreshData.refreshToken)

  return makeRequest(refreshData.accessToken)
}