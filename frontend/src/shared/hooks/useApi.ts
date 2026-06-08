import { useEffect, useState } from 'react'
import { message } from 'antd'

interface UseApiOptions<T> {
  onSuccess?: (data: T) => void
  onError?: (error: Error) => void
  showError?: boolean
}

export const useApi = <T,>(
  apiCall: () => Promise<T>,
  deps: any[] = [],
  options: UseApiOptions<T> = {}
) => {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    let isMounted = true

    const fetchData = async () => {
      try {
        setLoading(true)
        const result = await apiCall()
        if (isMounted) {
          setData(result)
          setError(null)
          options.onSuccess?.(result)
        }
      } catch (err) {
        if (isMounted) {
          const error = err instanceof Error ? err : new Error('Unknown error')
          setError(error)
          setData(null)
          options.onError?.(error)
          if (options.showError !== false) {
            message.error(error.message)
          }
        }
      } finally {
        if (isMounted) {
          setLoading(false)
        }
      }
    }

    fetchData()

    return () => {
      isMounted = false
    }
  }, deps)

  return { data, loading, error }
}

export const useApiMutation = <T, P>(
  apiCall: (params: P) => Promise<T>,
  options: UseApiOptions<T> = {}
) => {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const mutate = async (params: P) => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiCall(params)
      options.onSuccess?.(result)
      return result
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error')
      setError(error)
      options.onError?.(error)
      if (options.showError !== false) {
        message.error(error.message)
      }
      throw error
    } finally {
      setLoading(false)
    }
  }

  return { mutate, loading, error }
}
