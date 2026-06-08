import api from '@/shared/lib/api'
import type { FilterOptions } from '@/shared/types'

export const exportAPI = {
  exportMyActivities: async () => {
    const response = await api.get('/export/me', {
      responseType: 'blob',
    })
    return response.data
  },

  exportSummary: async (filters?: FilterOptions) => {
    const params = new URLSearchParams()
    if (filters?.group) params.append('group', filters.group)
    if (filters?.category) params.append('category', filters.category)
    
    const response = await api.get(`/export/summary?${params}`, {
      responseType: 'blob',
    })
    return response.data
  },
}
