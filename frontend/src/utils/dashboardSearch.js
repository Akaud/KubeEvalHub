export const dashboardSearchItems = [
  {
    id: 'dashboard-home',
    title: 'Overview',
    description: 'System overview and statistics',
    route: '/dashboard',
    keywords: [
      'dashboard',
      'home',
      'overview',
      'stats',
      'clusters',
      'agents',
    ],
  },

  {
    id: 'clusters-page',
    title: 'Clusters',
    description: 'Create clusters, assign agents, manage access, view metrics',
    route: '/dashboard/clusters',
    keywords: [
      'clusters',
      'cluster',
      'kubernetes',
      'inventory',
      'recommendations',
      'metrics',
      'access',
      'roles',
      'assign',
      'delete',
    ],
  },

  {
    id: 'agents-page',
    title: 'Agents',
    description: 'Create agents, install YAML, enable or disable ingestion',
    route: '/dashboard/agents',
    keywords: [
      'agents',
      'agent',
      'token',
      'yaml',
      'install',
      'enable',
      'disable',
      'delete',
      'scrape',
    ],
  },

  {
    id: 'settings-page',
    title: 'Account settings',
    description: 'Manage appearance, account information, and password',
    route: '/dashboard/settings',
    keywords: [
      'settings',
      'account',
      'theme',
      'dark',
      'light',
      'password',
      'email',
      'username',
      'profile',
      'security',
      'appearance',
    ],
  },

  {
    id: 'help-page',
    title: 'Help',
    description: 'Help and support page',
    route: '/dashboard/help',
    keywords: [
      'help',
      'support',
      'faq',
      'guide',
      'documentation',
    ],
  },
]

export function searchDashboardItems(query, items = dashboardSearchItems) {
  const normalizedQuery = query.trim().toLowerCase()

  if (!normalizedQuery) return []

  return items
    .map((item) => {
      const title = item.title.toLowerCase()
      const description = item.description.toLowerCase()
      const keywords = (item.keywords || []).map((k) => k.toLowerCase())

      let score = 0

      // strongest signal
      if (title.startsWith(normalizedQuery)) score += 5
      if (title.includes(normalizedQuery)) score += 3

      // keywords
      if (keywords.some((k) => k.includes(normalizedQuery))) score += 2

      // description
      if (description.includes(normalizedQuery)) score += 1

      if (score === 0) return null

      return { ...item, score }
    })
    .filter(Boolean)
    .sort((a, b) => b.score - a.score)
}