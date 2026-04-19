import axios from 'axios'
import { useAuthStore } from '@/store/auth'

// TODO: Replace with proper toast system in later task
const message = {
  error: (msg: string) => console.error('[API Error]', msg),
}

const client = axios.create({
  baseURL: '/api/v1',
})

// Request interceptor: attach Authorization header
client.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response envelope types
export interface ApiSuccessEnvelope<T> {
  code: 0
  data: T
}

export interface ApiErrorEnvelope {
  code: string
  message: string
}

// Response interceptor: unwrap data and handle errors
client.interceptors.response.use(
  (response) => {
    const body = response.data
    if (body && body.code === 0) {
      return body.data
    }
    return response.data
  },
  (error) => {
    if (!error.response) {
      return Promise.reject(error)
    }

    const { status } = error.response

    switch (status) {
      case 401:
        useAuthStore.getState().clearAuth()
        window.location.href = '/login'
        break
      case 403:
        message.error('权限不足')
        useAuthStore.getState().fetchPermissions()
        break
      case 404:
        message.error('资源不存在')
        break
      case 422:
        // Re-throw for component-level handling
        break
      case 500:
        message.error('服务器错误，请稍后重试')
        break
    }

    return Promise.reject(error)
  },
)

export default client
