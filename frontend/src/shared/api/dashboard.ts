import api from '@/shared/lib/api'
import type { DashboardMe, StudentStats, FilterOptions } from '@/shared/types'

export const dashboardAPI = {
  getMyDashboard: async () => {
    const response = await api.get<DashboardMe>('/dashboard/me')
    return response.data
  },

  getSummary: async (filters?: FilterOptions) => {
    const params = new URLSearchParams()
    if (filters?.group) params.append('group', filters.group)
    if (filters?.category) params.append('category', filters.category)
    if (filters?.student_id) params.append('student_id', filters.student_id)
    
    const response = await api.get<StudentStats[]>(`/dashboard/summary?${params}`)
    return response.data
  },
}
