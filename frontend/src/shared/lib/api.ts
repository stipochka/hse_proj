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
  (error) => Promise.reject(error),
)

export default baseInstance
