import { describe, it, expect, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import { ToastProvider } from '@/components/ui/toast'

beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

describe('Permission-driven UI', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', {
      bizKey: '1',
      username: 'testuser',
      displayName: 'Test User',
      isSuperAdmin: false,
      createTime: '',
    })
    useTeamStore.getState().setTeams([])
    useTeamStore.getState().setCurrentTeam(null)
  })

  describe('Sidebar - view:gantt permission', () => {
    it('hides gantt nav item when user lacks view:gantt permission', async () => {
      useAuthStore.getState().setPermissions({ isSuperAdmin: false, teamPermissions: {} })
      const { default: Sidebar } = await import('@/components/layout/Sidebar')
      render(
        <MemoryRouter>
          <Routes>
            <Route path="*" element={<Sidebar />} />
          </Routes>
        </MemoryRouter>,
      )
      expect(screen.queryByText('整体进度')).not.toBeInTheDocument()
    })

    it('shows gantt nav item when user has view:gantt permission', async () => {
      useAuthStore.getState().setPermissions({
        isSuperAdmin: false,
        teamPermissions: { 1: ['view:gantt'] },
      })
      const { default: Sidebar } = await import('@/components/layout/Sidebar')
      render(
        <MemoryRouter>
          <Routes>
            <Route path="*" element={<Sidebar />} />
          </Routes>
        </MemoryRouter>,
      )
      expect(screen.getByText('整体进度')).toBeInTheDocument()
    })
  })

  describe('Sidebar - admin items use permission codes', () => {
    it('shows user management when user has user:list permission', async () => {
      useAuthStore.getState().setPermissions({
        isSuperAdmin: false,
        teamPermissions: { 1: ['user:list'] },
      })
      const { default: Sidebar } = await import('@/components/layout/Sidebar')
      render(
        <MemoryRouter>
          <Routes>
            <Route path="*" element={<Sidebar />} />
          </Routes>
        </MemoryRouter>,
      )
      expect(screen.getByText('用户管理')).toBeInTheDocument()
    })

    it('shows role management when user has role:read permission', async () => {
      useAuthStore.getState().setPermissions({
        isSuperAdmin: false,
        teamPermissions: { 1: ['role:read'] },
      })
      const { default: Sidebar } = await import('@/components/layout/Sidebar')
      render(
        <MemoryRouter>
          <Routes>
            <Route path="*" element={<Sidebar />} />
          </Routes>
        </MemoryRouter>,
      )
      expect(screen.getByText('角色管理')).toBeInTheDocument()
    })
  })

  describe('TeamManagementPage - team:create permission', () => {
    it('hides create team button without team:create permission', async () => {
      useAuthStore.getState().setPermissions({ isSuperAdmin: false, teamPermissions: {} })
      server.use(http.get('/v1/teams', () => HttpResponse.json({ code: 0, data: [] })))
      const { default: TeamManagementPage } = await import('@/pages/TeamManagementPage')
      render(
        <QueryClientProvider client={createQueryClient()}>
          <ToastProvider>
            <MemoryRouter>
              <TeamManagementPage />
            </MemoryRouter>
          </ToastProvider>
        </QueryClientProvider>,
      )
      expect(screen.queryByText('创建团队')).not.toBeInTheDocument()
    })

    it('shows create team button with team:create permission', async () => {
      useAuthStore.getState().setPermissions({
        isSuperAdmin: false,
        teamPermissions: { 1: ['team:create'] },
      })
      server.use(http.get('/v1/teams', () => HttpResponse.json({ code: 0, data: [] })))
      const { default: TeamManagementPage } = await import('@/pages/TeamManagementPage')
      render(
        <QueryClientProvider client={createQueryClient()}>
          <ToastProvider>
            <MemoryRouter>
              <TeamManagementPage />
            </MemoryRouter>
          </ToastProvider>
        </QueryClientProvider>,
      )
      expect(screen.getByText('创建团队')).toBeInTheDocument()
    })
  })

  describe('TeamDetailPage - permission-controlled buttons', () => {
    function setupHandlers() {
      server.use(
        http.get('/v1/teams/:teamId', () =>
          HttpResponse.json({
            code: 0, data: {
              bizKey: '1', name: 'Test Team', description: '', code: '', pmKey: '1', pmDisplayName: 'PM',
              memberCount: 2, mainItemCount: 0, createTime: '2026-01-01', dbUpdateTime: '2026-01-01',
            },
          })),
        http.get('/v1/teams/:teamId/members', () =>
          HttpResponse.json({
            code: 0, data: [
              { id: 1, bizKey: '1', teamKey: '1', userKey: 'U001', displayName: 'PM', username: 'pm', role: 'pm', roleId: 1, roleName: 'pm', joinedAt: '2026-01-01' },
              { id: 2, bizKey: '2', teamKey: '1', userKey: 'U002', displayName: 'Member', username: 'member', role: 'member', roleId: 2, roleName: 'member', joinedAt: '2026-01-01' },
            ],
          })),
      )
    }

    async function renderTeamDetail() {
      setupHandlers()
      const { default: TeamDetailPage } = await import('@/pages/TeamDetailPage')
      const { ToastProvider } = await import('@/components/ui/toast')
      return render(
        <QueryClientProvider client={createQueryClient()}>
          <ToastProvider>
            <MemoryRouter initialEntries={['/teams/1']}>
              <Routes>
                <Route path="/teams/:teamId" element={<TeamDetailPage />} />
              </Routes>
            </MemoryRouter>
          </ToastProvider>
        </QueryClientProvider>,
      )
    }

    it('hides invite button without team:invite permission', async () => {
      useAuthStore.getState().setPermissions({ isSuperAdmin: false, teamPermissions: { 1: [] } })
      await renderTeamDetail()
      await waitFor(() => expect(screen.getByText('成员列表')).toBeInTheDocument())
      expect(screen.queryByText('添加成员')).not.toBeInTheDocument()
    })

    it('hides dissolve button without team:delete permission', async () => {
      useAuthStore.getState().setPermissions({ isSuperAdmin: false, teamPermissions: { 1: [] } })
      await renderTeamDetail()
      await waitFor(() => expect(screen.getByText('成员列表')).toBeInTheDocument())
      // The "解散团队" label still appears, but the danger button should not
      const dangerButton = screen.queryByRole('button', { name: '解散团队' })
      expect(dangerButton).not.toBeInTheDocument()
    })

    it('shows invite button with team:invite permission', async () => {
      useAuthStore.getState().setPermissions({
        isSuperAdmin: false,
        teamPermissions: { 1: ['team:invite'] },
      })
      await renderTeamDetail()
      await waitFor(() => expect(screen.getByText('添加成员')).toBeInTheDocument())
    })
  })
})
