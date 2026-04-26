import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User, PermissionData } from '@/types'
import { getPermissionsApi } from '@/api/permissions'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  isSuperAdmin: boolean
  permissions: PermissionData | null
  permissionsLoadedAt: number | null
  _hasHydrated: boolean
  setAuth: (token: string, user: User) => void
  clearAuth: () => void
  setPermissions: (permissions: PermissionData) => void
  fetchPermissions: () => Promise<void>
  hasPermission: (code: string, teamId?: string) => boolean
  _setHasHydrated: (v: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      isSuperAdmin: false,
      permissions: null,
      permissionsLoadedAt: null,
      _hasHydrated: false,
      setAuth: (token, user) =>
        set({
          token,
          user,
          isAuthenticated: token !== null,
          isSuperAdmin: user?.isSuperAdmin === true,
        }),
      clearAuth: () =>
        set({
          token: null,
          user: null,
          isAuthenticated: false,
          isSuperAdmin: false,
          permissions: null,
          permissionsLoadedAt: null,
          _hasHydrated: true,
        }),
      setPermissions: (permissions) =>
        set({ permissions, permissionsLoadedAt: Date.now() }),
      fetchPermissions: async () => {
        try {
          const permissions = await getPermissionsApi()
          set({ permissions, permissionsLoadedAt: Date.now() })
        } catch {
          set({ permissions: null })
        }
      },
      hasPermission: (code, teamId) => {
        const { permissions } = get()
        if (!permissions) return false
        if (permissions.isSuperAdmin) return true
        if (teamId !== undefined) {
          return permissions.teamPermissions[teamId]?.includes(code) ?? false
        }
        return Object.values(permissions.teamPermissions).some((codes) =>
          codes.includes(code),
        )
      },
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
