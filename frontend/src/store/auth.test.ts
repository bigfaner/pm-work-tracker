import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore } from './auth'
import type { User } from '@/types'

const mockUser: User = {
  id: 1,
  username: 'testuser',
  displayName: 'Test User',
  isSuperAdmin: false,
  canCreateTeam: false,
}

const superAdminUser: User = {
  ...mockUser,
  id: 2,
  username: 'admin',
  isSuperAdmin: true,
}

describe('useAuthStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.getState().clearAuth()
  })

  it('starts with null token and user', () => {
    const state = useAuthStore.getState()
    expect(state.token).toBeNull()
    expect(state.user).toBeNull()
  })

  it('setAuth populates token and user', () => {
    useAuthStore.getState().setAuth('jwt-token-123', mockUser)

    const state = useAuthStore.getState()
    expect(state.token).toBe('jwt-token-123')
    expect(state.user).toEqual(mockUser)
  })

  it('clearAuth resets token and user to null', () => {
    useAuthStore.getState().setAuth('jwt-token-123', mockUser)
    useAuthStore.getState().clearAuth()

    const state = useAuthStore.getState()
    expect(state.token).toBeNull()
    expect(state.user).toBeNull()
  })

  it('isAuthenticated returns false when no token', () => {
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })

  it('isAuthenticated returns true when token is set', () => {
    useAuthStore.getState().setAuth('jwt-token-123', mockUser)
    expect(useAuthStore.getState().isAuthenticated).toBe(true)
  })

  it('isAuthenticated returns false after clearAuth', () => {
    useAuthStore.getState().setAuth('jwt-token-123', mockUser)
    useAuthStore.getState().clearAuth()
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })

  it('isSuperAdmin returns false when no user', () => {
    expect(useAuthStore.getState().isSuperAdmin).toBe(false)
  })

  it('isSuperAdmin returns false for regular user', () => {
    useAuthStore.getState().setAuth('jwt-token-123', mockUser)
    expect(useAuthStore.getState().isSuperAdmin).toBe(false)
  })

  it('isSuperAdmin returns true for super admin user', () => {
    useAuthStore.getState().setAuth('jwt-token-456', superAdminUser)
    expect(useAuthStore.getState().isSuperAdmin).toBe(true)
  })
})
