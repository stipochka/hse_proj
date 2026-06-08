// DEPRECATED: This file is kept for backward compatibility but these endpoints are no longer used.
// Student data is now retrieved through dashboard summary API:
// - Use dashboardAPI.getSummary() to get StudentStats (which includes student groups and stats)

import api from '@/shared/lib/api'
import type { Student, FilterOptions } from '@/shared/types'

/**
 * @deprecated Use dashboardAPI.getSummary() instead
 */
export const studentsAPI = {
  getGroupStudents: async (groupId: string, page = 1, pageSize = 20) => {
    const params = { page, pageSize }
    const response = await api.get<{
      data: Student[]
      total: number
      page: number
      pageSize: number
    }>(`/groups/${groupId}/students`, { params })
    return response.data
  },

  /**
   * @deprecated Use dashboardAPI.getSummary() instead
   */
  getStudents: async (filters?: FilterOptions, page = 1, pageSize = 20) => {
    const params = {
      page,
      pageSize,
      ...filters,
    }
    const response = await api.get<{
      data: Student[]
      total: number
      page: number
      pageSize: number
    }>('/students', { params })
    return response.data
  },

  /**
   * @deprecated Use activitiesAPI.getActivityById() instead
   */
  getStudentById: async (id: string) => {
    const response = await api.get<Student>(`/students/${id}`)
    return response.data
  },
}
