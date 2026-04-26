import { describe, it, expect, vi, beforeEach } from 'vitest'
import client from './client'
import { useAuthStore } from '@/store/auth'
import { showToast } from '@/lib/toast'

vi.mock('@/lib/toast', () => ({
  showToast: vi.fn(),
}))

describe('API Client', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    vi.clearAllMocks()
    // Reset window.location.href mock
    const originalLocation = window.location
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...originalLocation, href: '' },
    })
  })

  describe('base configuration', () => {
    it('should have baseURL set to /v1', () => {
      expect(client.defaults.baseURL).toBe('/v1')
    })
  })

  describe('request interceptor', () => {
    it('should not attach Authorization header when no token', () => {
      const handlers = (client.interceptors.request as any).handlers
      const fulfilled = handlers[handlers.length - 1].fulfilled
      const config = { headers: {} as Record<string, string> }
      const result = fulfilled(config)
      expect(result.headers.Authorization).toBeUndefined()
    })

    it('should attach Authorization header when token exists in auth store', () => {
      useAuthStore.getState().setAuth('test-jwt-token', {
        bizKey: '1',
        username: 'testuser',
        displayName: 'Test User',
        isSuperAdmin: false,
        createTime: '2026-01-01T00:00:00Z',
      })

      // Access the request interceptor directly through client
      const handlers = (client.interceptors.request as any).handlers
      const fulfilled = handlers[handlers.length - 1].fulfilled
      const config = { headers: {} as Record<string, string> }
      const result = fulfilled(config)
      expect(result.headers.Authorization).toBe('Bearer test-jwt-token')
    })

    it('should not attach Authorization header when token is null', () => {
      const handlers = (client.interceptors.request as any).handlers
      const fulfilled = handlers[handlers.length - 1].fulfilled
      const config = { headers: {} as Record<string, string> }
      const result = fulfilled(config)
      expect(result.headers.Authorization).toBeUndefined()
    })
  })

  describe('response interceptor - success', () => {
    it('should unwrap response.data.data for success responses', () => {
      const handlers = (client.interceptors.response as any).handlers
      const fulfilled = handlers[handlers.length - 1].fulfilled

      const response = {
        data: { code: 0, data: { id: 1, name: 'test' } },
      }
      const result = fulfilled(response)
      expect(result).toEqual({ id: 1, name: 'test' })
    })

    it('should pass through response.data if code is not 0', () => {
      const handlers = (client.interceptors.response as any).handlers
      const fulfilled = handlers[handlers.length - 1].fulfilled

      const response = {
        data: { code: 999, data: 'something' },
      }
      const result = fulfilled(response)
      expect(result).toEqual({ code: 999, data: 'something' })
    })
  })

  describe('response interceptor - errors', () => {
    function getRejectHandler() {
      const handlers = (client.interceptors.response as any).handlers
      return handlers[handlers.length - 1].rejected
    }

    it('should clear auth and redirect on 401', async () => {
      useAuthStore.getState().setAuth('token', {
        bizKey: '1',
        username: 'u',
        displayName: 'U',
        isSuperAdmin: false,
        createTime: '',
      })

      const handler = getRejectHandler()
      const error = { response: { status: 401, data: {} } }

      await expect(handler(error)).rejects.toBe(error)
      expect(useAuthStore.getState().token).toBeNull()
      expect(window.location.href).toBe('/login')
    })

    it('should log error on 403', async () => {
      const handler = getRejectHandler()
      const error = { response: { status: 403, data: {} } }

      await expect(handler(error)).rejects.toBe(error)
      expect(showToast).toHaveBeenCalledWith('权限不足', 'error')
    })

    it('should log error on 404', async () => {
      const handler = getRejectHandler()
      const error = { response: { status: 404, data: {} } }

      await expect(handler(error)).rejects.toBe(error)
      expect(showToast).toHaveBeenCalledWith('资源不存在', 'error')
    })

    it('should re-throw 422 errors without logging message', async () => {
      const handler = getRejectHandler()
      const error = { response: { status: 422, data: { code: 'VALIDATION_ERROR', message: 'bad' } } }

      await expect(handler(error)).rejects.toBe(error)
      expect(showToast).not.toHaveBeenCalled()
    })

    it('should log error on 500', async () => {
      const handler = getRejectHandler()
      const error = { response: { status: 500, data: {} } }

      await expect(handler(error)).rejects.toBe(error)
      expect(showToast).toHaveBeenCalledWith('服务器错误，请稍后重试', 'error')
    })

    it('should pass through network errors without response', async () => {
      const handler = getRejectHandler()
      const error = new Error('Network Error')

      await expect(handler(error)).rejects.toThrow('Network Error')
      expect(showToast).not.toHaveBeenCalled()
    })
  })
})
