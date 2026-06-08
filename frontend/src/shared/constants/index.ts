export const APP_CONFIG = {
  NAME: 'HSE Admin',
  VERSION: '0.1.0',
}

export const API_ROUTES = {
  ACTIVITIES: '/activities',
  ACTIVITIES_MY: '/activities/my',
  DASHBOARD_ME: '/dashboard/me',
  DASHBOARD_SUMMARY: '/dashboard/summary',
  EXPORT_ME: '/export/me',
  EXPORT_SUMMARY: '/export/summary',
}

export const PAGE_ROUTES = {
  HOME: '/',
  ACTIVITIES: '/activities',
  ACTIVITY_EVALUATE: '/activities/:id/evaluate',
  EXPORT: '/export',
}

export const ACTIVITY_STATUS = {
  PENDING: 'PENDING',
  SUBMITTED: 'SUBMITTED',
  EVALUATED: 'EVALUATED',
  REJECTED: 'REJECTED',
} as const

export const ACTIVITY_STATUS_LABELS: Record<string, string> = {
  PENDING: 'В ожидании',
  SUBMITTED: 'На проверку',
  EVALUATED: 'Проверено',
  REJECTED: 'Отклонено',
}

export const DEFAULT_PAGE_SIZE = 20
export const DEFAULT_TABLE_PAGE_SIZE_OPTIONS = [10, 20, 50, 100]
export const MAX_FILE_SIZE = 50 * 1024 * 1024 // 50MB
export const ALLOWED_FILE_TYPES = ['.pdf']

export const DEBOUNCE_DELAY = 300
export const THROTTLE_DELAY = 1000
