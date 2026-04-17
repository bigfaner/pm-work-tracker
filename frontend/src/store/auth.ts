import { create } from 'zustand'
import type { User } from '@/types'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  isSuperAdmin: boolean
  setAuth: (token: string, user: User) => void
  clearAuth: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  isAuthenticated: false,
  isSuperAdmin: false,
  setAuth: (token, user) =>
    set({
      token,
      user,
      isAuthenticated: token !== null,
      isSuperAdmin: user?.is_super_admin === true,
    }),
  clearAuth: () =>
    set({ token: null, user: null, isAuthenticated: false, isSuperAdmin: false }),
}))
