export function getStoredTheme() {
  return localStorage.getItem('theme') || 'light'
}

export function applyTheme(theme) {
  document.body.classList.remove('theme-light', 'theme-dark')
  document.body.classList.add(theme === 'dark' ? 'theme-dark' : 'theme-light')
  localStorage.setItem('theme', theme)
  window.dispatchEvent(new Event('theme-change'))
}