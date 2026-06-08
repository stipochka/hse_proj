export const formatDate = (date: string | Date, format: 'short' | 'long' = 'short'): string => {
  const d = typeof date === 'string' ? new Date(date) : date

  if (format === 'short') {
    return d.toLocaleDateString('ru-RU')
  }

  return d.toLocaleDateString('ru-RU', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export const formatTime = (date: string | Date): string => {
  const d = typeof date === 'string' ? new Date(date) : date
  return d.toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export const getStatusColor = (status: string): string => {
  const statusColorMap: Record<string, string> = {
    SUBMITTED: 'orange',
    EVALUATED: 'green',
    REJECTED: 'red',
    DRAFT: 'default',
    ACTIVE: 'green',
    INACTIVE: 'gray',
    PENDING: 'orange',
    COMPLETED: 'green',
    FAILED: 'red',
  }

  return statusColorMap[status] || 'default'
}

export const getStatusLabel = (status: string): string => {
  const statusLabelMap: Record<string, string> = {
    SUBMITTED: 'На проверку',
    EVALUATED: 'Проверено',
    REJECTED: 'Отклонено',
    DRAFT: 'Черновик',
    ACTIVE: 'Активно',
    INACTIVE: 'Неактивно',
    PENDING: 'Ожидание',
    COMPLETED: 'Завершено',
    FAILED: 'Ошибка',
  }

  return statusLabelMap[status] || status
}

export const downloadFile = (blob: Blob, filename: string) => {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout | null = null

  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean = false

  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => {
        inThrottle = false
      }, limit)
    }
  }
}
