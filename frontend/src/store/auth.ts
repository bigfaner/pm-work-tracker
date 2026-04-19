import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  isSuperAdmin: boolean
  _hasHydrated: boolean
  setAuth: (token: string, user: any) => void
  clearAuth: () => void
  _setHasHydrated: (v: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      isSuperAdmin: false,
      _hasHydrated: false,
      setAuth: (token, user) =>
        set({
          token,
          user,
          isAuthenticated: token !== null,
          isSuperAdmin: (user?.isSuperAdmin ?? user?.isSuperAdmin) === true,
        }),
      clearAuth: () =>
        set({ token: null, user: null, isAuthenticated: false, isSuperAdmin: false, _hasHydrated: true }),
      _setHasHydrated: (v) => set({ _hasHydrated: v }),
    }),
    {
      name: 'auth-storage',
      onRehydrateStorage: () => (state) => {
        state?._setHasHydrated(true)
      },
    },
  ),
)
