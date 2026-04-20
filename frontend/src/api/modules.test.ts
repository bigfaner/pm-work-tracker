import { describe, it, expect, vi, beforeEach } from 'vitest'
import client from './client'
import * as authApi from './auth'
import * as teamsApi from './teams'
import * as mainItemsApi from './mainItems'
import * as subItemsApi from './subItems'
import * as progressApi from './progress'
import * as itemPoolApi from './itemPool'
import * as viewsApi from './views'
import * as reportsApi from './reports'
import * as adminApi from './admin'

vi.mock('./client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    patch: vi.fn(),
  },
}))

const mockClient = vi.mocked(client, true)

describe('API modules', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('auth', () => {
    it('loginApi should POST to /auth/login', async () => {
      mockClient.post.mockResolvedValue({ token: 't', user: {} })
      await authApi.loginApi({ username: 'u', password: 'p' })
      expect(mockClient.post).toHaveBeenCalledWith('/auth/login', { username: 'u', password: 'p' })
    })

    it('logoutApi should POST to /auth/logout', async () => {
      mockClient.post.mockResolvedValue(undefined)
      await authApi.logoutApi()
      expect(mockClient.post).toHaveBeenCalledWith('/auth/logout')
    })
  })

  describe('teams', () => {
    it('createTeamApi should POST to /teams', async () => {
      mockClient.post.mockResolvedValue({ id: 1 })
      await teamsApi.createTeamApi({ name: 'Team' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams', { name: 'Team' })
    })

    it('listTeamsApi should GET /teams', async () => {
      mockClient.get.mockResolvedValue([])
      await teamsApi.listTeamsApi()
      expect(mockClient.get).toHaveBeenCalledWith('/teams')
    })

    it('getTeamApi should GET /teams/:id', async () => {
      mockClient.get.mockResolvedValue({})
      await teamsApi.getTeamApi(5)
      expect(mockClient.get).toHaveBeenCalledWith('/teams/5')
    })

    it('updateTeamApi should PUT /teams/:id', async () => {
      mockClient.put.mockResolvedValue({})
      await teamsApi.updateTeamApi(5, { name: 'New' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/5', { name: 'New' })
    })

    it('deleteTeamApi should DELETE /teams/:id with body', async () => {
      mockClient.delete.mockResolvedValue(undefined)
      await teamsApi.deleteTeamApi(5, { confirmName: 'Team' })
      expect(mockClient.delete).toHaveBeenCalledWith('/teams/5', { data: { confirmName: 'Team' } })
    })

    it('listMembersApi should GET /teams/:id/members', async () => {
      mockClient.get.mockResolvedValue([])
      await teamsApi.listMembersApi(5)
      expect(mockClient.get).toHaveBeenCalledWith('/teams/5/members')
    })

    it('inviteMemberApi should POST to /teams/:id/members', async () => {
      mockClient.post.mockResolvedValue(undefined)
      await teamsApi.inviteMemberApi(5, { username: 'u', roleId: 3 })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/5/members', { username: 'u', roleId: 3 })
    })

    it('removeMemberApi should DELETE /teams/:teamId/members/:userId', async () => {
      mockClient.delete.mockResolvedValue(undefined)
      await teamsApi.removeMemberApi(5, 10)
      expect(mockClient.delete).toHaveBeenCalledWith('/teams/5/members/10')
    })

    it('transferPmApi should PUT /teams/:id/pm', async () => {
      mockClient.put.mockResolvedValue(undefined)
      await teamsApi.transferPmApi(5, { newPmUserId: 3 })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/5/pm', { newPmUserId: 3 })
    })
  })

  describe('mainItems', () => {
    it('createMainItemApi should POST /teams/:id/main-items', async () => {
      mockClient.post.mockResolvedValue({})
      await mainItemsApi.createMainItemApi(1, { title: 'Item', priority: 'P0', assigneeId: 0, startDate: '', expectedEndDate: '' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/1/main-items', { title: 'Item', priority: 'P0', assigneeId: 0, startDate: '', expectedEndDate: '' })
    })

    it('listMainItemsApi should GET /teams/:id/main-items with params', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await mainItemsApi.listMainItemsApi(1, { status: 'open' })
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/main-items', { params: { status: 'open' } })
    })

    it('getMainItemApi should GET /teams/:id/main-items/:itemId', async () => {
      mockClient.get.mockResolvedValue({})
      await mainItemsApi.getMainItemApi(1, 2)
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/main-items/2')
    })

    it('updateMainItemApi should PUT /teams/:id/main-items/:itemId', async () => {
      mockClient.put.mockResolvedValue({})
      await mainItemsApi.updateMainItemApi(1, 2, { title: 'Updated' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/1/main-items/2', { title: 'Updated' })
    })

    it('archiveMainItemApi should POST /teams/:id/main-items/:itemId/archive', async () => {
      mockClient.post.mockResolvedValue(undefined)
      await mainItemsApi.archiveMainItemApi(1, 2)
      expect(mockClient.post).toHaveBeenCalledWith('/teams/1/main-items/2/archive')
    })
  })

  describe('subItems', () => {
    it('createSubItemApi should POST to correct URL', async () => {
      mockClient.post.mockResolvedValue({})
      await subItemsApi.createSubItemApi(1, 2, { title: 'Sub', priority: 'P1', assigneeId: 3, startDate: '', expectedEndDate: '' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/1/main-items/2/sub-items', { mainItemId: 2, title: 'Sub', priority: 'P1', assigneeId: 3, startDate: '', expectedEndDate: '' })
    })

    it('listSubItemsApi should GET with params', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await subItemsApi.listSubItemsApi(1, 2, { status: 'open' })
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/main-items/2/sub-items', { params: { status: 'open' } })
    })

    it('getSubItemApi should GET /teams/:id/sub-items/:itemId', async () => {
      mockClient.get.mockResolvedValue({})
      await subItemsApi.getSubItemApi(1, 3)
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/sub-items/3')
    })

    it('updateSubItemApi should PUT /teams/:id/sub-items/:itemId', async () => {
      mockClient.put.mockResolvedValue({})
      await subItemsApi.updateSubItemApi(1, 3, { title: 'Updated' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/1/sub-items/3', { title: 'Updated' })
    })

    it('changeSubItemStatusApi should PUT to status endpoint', async () => {
      mockClient.put.mockResolvedValue(undefined)
      await subItemsApi.changeSubItemStatusApi(1, 3, { status: 'done' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/1/sub-items/3/status', { status: 'done' })
    })

    it('assignSubItemApi should PUT to assignee endpoint', async () => {
      mockClient.put.mockResolvedValue(undefined)
      await subItemsApi.assignSubItemApi(1, 3, { assigneeId: 5 })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/1/sub-items/3/assignee', { assigneeId: 5 })
    })
  })

  describe('progress', () => {
    it('appendProgressApi should POST to progress endpoint', async () => {
      mockClient.post.mockResolvedValue({})
      await progressApi.appendProgressApi(1, 2, { completion: 80 })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/1/sub-items/2/progress', { completion: 80 })
    })

    it('listProgressApi should GET progress list', async () => {
      mockClient.get.mockResolvedValue([])
      await progressApi.listProgressApi(1, 2)
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/sub-items/2/progress')
    })

    it('correctCompletionApi should PATCH completion endpoint', async () => {
      mockClient.patch.mockResolvedValue({})
      await progressApi.correctCompletionApi(1, 5, { completion: 90 })
      expect(mockClient.patch).toHaveBeenCalledWith('/teams/1/progress/5/completion', { completion: 90 })
    })
  })

  describe('itemPool', () => {
    it('submitItemPoolApi should POST to item-pool', async () => {
      mockClient.post.mockResolvedValue({})
      await itemPoolApi.submitItemPoolApi(1, { title: 'Pool Item' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/1/item-pool', { title: 'Pool Item' })
    })

    it('listItemPoolApi should GET with params', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await itemPoolApi.listItemPoolApi(1, { status: 'open' })
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/item-pool', { params: { status: 'open' } })
    })

    it('getItemPoolApi should GET /teams/:id/item-pool/:poolId', async () => {
      mockClient.get.mockResolvedValue({})
      await itemPoolApi.getItemPoolApi(1, 3)
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/item-pool/3')
    })

    it('assignItemPoolApi should POST to assign endpoint', async () => {
      mockClient.post.mockResolvedValue({ subItemId: 10 })
      await itemPoolApi.assignItemPoolApi(1, 3, { mainItemId: 2, assigneeId: 5, priority: 'P2', startDate: '', expectedEndDate: '' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/1/item-pool/3/assign', { mainItemId: 2, assigneeId: 5, priority: 'P2', startDate: '', expectedEndDate: '' })
    })

    it('rejectItemPoolApi should POST to reject endpoint', async () => {
      mockClient.post.mockResolvedValue({})
      await itemPoolApi.rejectItemPoolApi(1, 3, { reason: 'bad' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/1/item-pool/3/reject', { reason: 'bad' })
    })
  })

  describe('views', () => {
    it('getWeeklyViewApi should GET weekly view', async () => {
      mockClient.get.mockResolvedValue({})
      await viewsApi.getWeeklyViewApi(1, '2026-01-06')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/views/weekly', { params: { weekStart: '2026-01-06' } })
    })

    it('getGanttViewApi should GET gantt view', async () => {
      mockClient.get.mockResolvedValue({})
      await viewsApi.getGanttViewApi(1, 'open')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/views/gantt', { params: { status: 'open' } })
    })

    it('getTableViewApi should GET table view with filter', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await viewsApi.getTableViewApi(1, { priority: 'P0' })
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/views/table', { params: { priority: 'P0' } })
    })

    it('exportTableCsvApi should GET table export as blob', async () => {
      mockClient.get.mockResolvedValue(new Blob())
      await viewsApi.exportTableCsvApi(1)
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/views/table/export', { params: undefined, responseType: 'blob' })
    })
  })

  describe('reports', () => {
    it('getWeeklyReportPreviewApi should GET preview', async () => {
      mockClient.get.mockResolvedValue({})
      await reportsApi.getWeeklyReportPreviewApi(1, '2026-01-06')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/reports/weekly/preview', { params: { weekStart: '2026-01-06' } })
    })

    it('exportWeeklyReportApi should GET export as blob', async () => {
      mockClient.get.mockResolvedValue(new Blob())
      await reportsApi.exportWeeklyReportApi(1, '2026-01-06')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/reports/weekly/export', { params: { weekStart: '2026-01-06' }, responseType: 'blob' })
    })
  })

  describe('admin', () => {
    it('listUsersApi should GET /admin/users', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await adminApi.listUsersApi({ page: 1, pageSize: 20 })
      expect(mockClient.get).toHaveBeenCalledWith('/admin/users', { params: { page: 1, pageSize: 20 } })
    })


    it('listAdminTeamsApi should GET /admin/teams', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await adminApi.listAdminTeamsApi(2, 10)
      expect(mockClient.get).toHaveBeenCalledWith('/admin/teams', { params: { page: 2, pageSize: 10 } })
    })
  })
})
