export interface Activity {
  id: number
  student_id: string
  student_group: string
  title: string
  description: string
  category: string
  status: 'PENDING' | 'SUBMITTED' | 'EVALUATED' | 'REJECTED'
  created_at: string
  evaluation?: Evaluation
}

export interface Evaluation {
  id: number
  activity_id: number
  admin_id: string
  points: number
  credits?: number
  comment?: string
  evaluated_at: string
}

export interface DashboardMe {
  activity_count: number
  total_points: number
  total_credits: number
  by_status: Record<string, number>
  by_category: Record<string, number>
}

export interface StudentStats {
  student_id: string
  student_group: string
  activity_count: number
  evaluated_count: number
  total_points: number
  total_credits: number
}

export interface UploadURLRequest {
  title: string
  description: string
  category: string
}

export interface UploadURLResponse {
  activity_id: number
  upload_url: string
  pdf_key: string
}

export interface FileURLResponse {
  file_url: string
}

export interface EvaluationRequest {
  points: number
  credits?: number
  comment?: string
  reject?: boolean
}

export interface FilterOptions {
  status?: string
  category?: string
  student_id?: string
  group?: string
}

// Types for deprecated APIs (kept for backward compatibility)
export interface User {
  id: string
  username: string
  email?: string
  firstName?: string
  lastName?: string
  roles?: string[]
}

export interface Student {
  id: string
  name: string
  email: string
  groupName: string
  totalScore: number
  activitiesCount: number
}

export interface Group {
  id: string
  name: string
}

export interface Course {
  id: string
  name: string
}

export interface Category {
  id: string
  name: string
}

