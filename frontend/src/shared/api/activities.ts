import api from '@/shared/lib/api'
import type {
  Activity,
  UploadURLRequest,
  UploadURLResponse,
  FileURLResponse,
  EvaluationRequest,
  Evaluation,
  FilterOptions,
} from '@/shared/types'

export const activitiesAPI = {
  // Admin endpoints
  getActivities: async (filters?: FilterOptions) => {
    const params = new URLSearchParams()
    if (filters?.group) params.append('group', filters.group)
    if (filters?.status) params.append('status', filters.status)
    if (filters?.category) params.append('category', filters.category)
    if (filters?.student_id) params.append('student_id', filters.student_id)
    
    const response = await api.get<Activity[]>(`/activities?${params}`)
    return response.data
  },

  // Student endpoints
  getMyActivities: async (filters?: FilterOptions) => {
    const params = new URLSearchParams()
    if (filters?.status) params.append('status', filters.status)
    if (filters?.category) params.append('category', filters.category)
    
    const response = await api.get<Activity[]>(`/activities/my?${params}`)
    return response.data
  },

  getActivityById: async (id: number) => {
    const response = await api.get<Activity>(`/activities/${id}`)
    return response.data
  },

  getUploadUrl: async (request: UploadURLRequest) => {
    const response = await api.post<UploadURLResponse>('/activities/upload-url', request)
    return response.data
  },

  confirmUpload: async (activityId: number) => {
    const response = await api.post<Activity>(`/activities/${activityId}/confirm`)
    return response.data
  },

  getFileUrl: async (activityId: number) => {
    const response = await api.get<FileURLResponse>(`/activities/${activityId}/file`)
    return response.data
  },

  evaluateActivity: async (activityId: number, evaluation: EvaluationRequest) => {
    const response = await api.post<Evaluation>(
      `/activities/${activityId}/evaluation`,
      evaluation
    )
    return response.data
  },
}

