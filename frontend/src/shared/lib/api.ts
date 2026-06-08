import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios'

let token: string | null = null

export const setToken = (newToken: string | null) => {
  token = newToken
}

const baseInstance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
  timeout: 10000,
})

baseInstance.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

baseInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)

export default baseInstance
