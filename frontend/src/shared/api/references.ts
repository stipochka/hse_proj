// DEPRECATED: This file is kept for backward compatibility but these endpoints are no longer used.
// Groups are now extracted from dashboard summary data.
// Categories and other references are handled by the backend.

import api from '@/shared/lib/api'
import type { Group, Course, Category } from '@/shared/types'

export interface Stream {
  id: string
  name: string
  description?: string
}

/**
 * @deprecated Use dashboardAPI.getSummary() to get students and groups
 */
export const referencesAPI = {
  getGroups: async () => {
    const response = await api.get<Group[]>('/groups')
    return response.data
  },

  getCourses: async () => {
    const response = await api.get<Course[]>('/courses')
    return response.data
  },

  getCategories: async () => {
    const response = await api.get<Category[]>('/categories')
    return response.data
  },

  getStreams: async () => {
    const response = await api.get<Stream[]>('/streams')
    return response.data
  },
}
