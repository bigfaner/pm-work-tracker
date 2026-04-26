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
      await teamsApi.createTeamApi({ name: 'Team', code: 'TEAM' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams', { name: 'Team', code: 'TEAM' })
    })

    it('listTeamsApi should GET /teams', async () => {
      mockClient.get.mockResolvedValue([])
      await teamsApi.listTeamsApi()
      expect(mockClient.get).toHaveBeenCalledWith('/teams', { params: undefined })
    })

    it('getTeamApi should GET /teams/:bizKey', async () => {
      mockClient.get.mockResolvedValue({})
      await teamsApi.getTeamApi('team-bk')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk')
    })

    it('updateTeamApi should PUT /teams/:bizKey', async () => {
      mockClient.put.mockResolvedValue({})
      await teamsApi.updateTeamApi('team-bk', { name: 'New' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/team-bk', { name: 'New' })
    })

    it('deleteTeamApi should DELETE /teams/:bizKey with body', async () => {
      mockClient.delete.mockResolvedValue(undefined)
      await teamsApi.deleteTeamApi('team-bk', { confirmName: 'Team' })
      expect(mockClient.delete).toHaveBeenCalledWith('/teams/team-bk', { data: { confirmName: 'Team' } })
    })

    it('listMembersApi should GET /teams/:bizKey/members', async () => {
      mockClient.get.mockResolvedValue([])
      await teamsApi.listMembersApi('team-bk')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk/members')
    })

    it('inviteMemberApi should POST to /teams/:bizKey/members', async () => {
      mockClient.post.mockResolvedValue(undefined)
      await teamsApi.inviteMemberApi('team-bk', { username: 'u', roleKey: 'member' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/team-bk/members', { username: 'u', roleKey: 'member' })
    })

    it('removeMemberApi should DELETE /teams/:bizKey/members/:userKey', async () => {
      mockClient.delete.mockResolvedValue(undefined)
      await teamsApi.removeMemberApi('team-bk', 'user-10')
      expect(mockClient.delete).toHaveBeenCalledWith('/teams/team-bk/members/user-10')
    })

    it('transferPmApi should PUT /teams/:bizKey/pm', async () => {
      mockClient.put.mockResolvedValue(undefined)
      await teamsApi.transferPmApi('team-bk', { newPmUserKey: 'user-3' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/team-bk/pm', { newPmUserKey: 'user-3' })
    })
  })

  describe('mainItems', () => {
    it('createMainItemApi should POST /teams/:teamBizKey/main-items', async () => {
      mockClient.post.mockResolvedValue({})
      await mainItemsApi.createMainItemApi('team-bk', { title: 'Item', priority: 'P0', assigneeKey: '', startDate: '', expectedEndDate: '' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/team-bk/main-items', { title: 'Item', priority: 'P0', assigneeKey: '', startDate: '', expectedEndDate: '' })
    })

    it('listMainItemsApi should GET /teams/:teamBizKey/main-items with params', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await mainItemsApi.listMainItemsApi('team-bk', { status: 'open' })
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk/main-items', { params: { status: 'open' } })
    })

    it('getMainItemApi should GET /teams/:teamBizKey/main-items/:bizKey', async () => {
      mockClient.get.mockResolvedValue({})
      await mainItemsApi.getMainItemApi('team-bk', 'item-bk')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk/main-items/item-bk')
    })

    it('updateMainItemApi should PUT /teams/:teamBizKey/main-items/:bizKey', async () => {
      mockClient.put.mockResolvedValue({})
      await mainItemsApi.updateMainItemApi('team-bk', 'item-bk', { title: 'Updated' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/team-bk/main-items/item-bk', { title: 'Updated' })
    })

    it('changeMainItemStatusApi should PUT /teams/:teamBizKey/main-items/:bizKey/status', async () => {
      mockClient.put.mockResolvedValue({ status: 'completed' })
      await mainItemsApi.changeMainItemStatusApi('team-bk', 'item-bk', { status: 'completed' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/team-bk/main-items/item-bk/status', { status: 'completed' })
    })

    it('getMainItemTransitionsApi should GET /teams/:teamBizKey/main-items/:bizKey/available-transitions', async () => {
      mockClient.get.mockResolvedValue(['progressing', 'closed'])
      await mainItemsApi.getMainItemTransitionsApi('team-bk', 'item-bk')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk/main-items/item-bk/available-transitions')
    })
  })

  describe('subItems', () => {
    it('createSubItemApi should POST to correct URL', async () => {
      mockClient.post.mockResolvedValue({})
      await subItemsApi.createSubItemApi('team-bk', 'main-bk', { title: 'Sub', priority: 'P1', assigneeKey: 'user-3', startDate: '', expectedEndDate: '' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/team-bk/main-items/main-bk/sub-items', { mainItemBizKey: 'main-bk', title: 'Sub', priority: 'P1', assigneeKey: 'user-3', startDate: '', expectedEndDate: '' })
    })

    it('listSubItemsApi should GET with params', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await subItemsApi.listSubItemsApi('team-bk', 'main-bk', { status: 'open' })
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk/main-items/main-bk/sub-items', { params: { status: 'open' } })
    })

    it('getSubItemApi should GET /teams/:teamBizKey/sub-items/:bizKey', async () => {
      mockClient.get.mockResolvedValue({})
      await subItemsApi.getSubItemApi('team-bk', 'sub-bk')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk/sub-items/sub-bk')
    })

    it('updateSubItemApi should PUT /teams/:teamBizKey/sub-items/:bizKey', async () => {
      mockClient.put.mockResolvedValue({})
      await subItemsApi.updateSubItemApi('team-bk', 'sub-bk', { title: 'Updated' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/team-bk/sub-items/sub-bk', { title: 'Updated' })
    })

    it('changeSubItemStatusApi should PUT to status endpoint', async () => {
      mockClient.put.mockResolvedValue(undefined)
      await subItemsApi.changeSubItemStatusApi('team-bk', 'sub-bk', { status: 'done' })
      expect(mockClient.put).toHaveBeenCalledWith('/teams/team-bk/sub-items/sub-bk/status', { status: 'done' })
    })

    it('getSubItemTransitionsApi should GET /teams/:teamBizKey/sub-items/:bizKey/available-transitions', async () => {
      mockClient.get.mockResolvedValue(['progressing', 'completed'])
      await subItemsApi.getSubItemTransitionsApi('team-bk', 'sub-bk')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk/sub-items/sub-bk/available-transitions')
    })
  })

  describe('progress', () => {
    it('appendProgressApi should POST to progress endpoint', async () => {
      mockClient.post.mockResolvedValue({})
      await progressApi.appendProgressApi('team-bk', 'sub-bk', { completion: 80 })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/team-bk/sub-items/sub-bk/progress', { completion: 80 })
    })

    it('listProgressApi should GET progress list', async () => {
      mockClient.get.mockResolvedValue([])
      await progressApi.listProgressApi('team-bk', 'sub-bk')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/team-bk/sub-items/sub-bk/progress')
    })

  })

  describe('itemPool', () => {
    it('submitItemPoolApi should POST to item-pool', async () => {
      mockClient.post.mockResolvedValue({})
      await itemPoolApi.submitItemPoolApi('1', { title: 'Pool Item' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/1/item-pool', { title: 'Pool Item' })
    })

    it('listItemPoolApi should GET with params', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await itemPoolApi.listItemPoolApi('1', { status: 'open' })
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/item-pool', { params: { status: 'open' } })
    })

    it('assignItemPoolApi should POST to assign endpoint', async () => {
      mockClient.post.mockResolvedValue({ mainItemBizKey: 'mi-2', subItemBizKey: 'si-10' })
      await itemPoolApi.assignItemPoolApi('team-bk', 'pool-bk', { mainItemKey: 'mi-2', assigneeKey: 'user-5', priority: 'P2', startDate: '', expectedEndDate: '' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/team-bk/item-pool/pool-bk/assign', { mainItemKey: 'mi-2', assigneeKey: 'user-5', priority: 'P2', startDate: '', expectedEndDate: '' })
    })

    it('rejectItemPoolApi should POST to reject endpoint', async () => {
      mockClient.post.mockResolvedValue({})
      await itemPoolApi.rejectItemPoolApi('team-bk', 'pool-bk', { reason: 'bad' })
      expect(mockClient.post).toHaveBeenCalledWith('/teams/team-bk/item-pool/pool-bk/reject', { reason: 'bad' })
    })
  })

  describe('views', () => {
    it('getWeeklyViewApi should GET weekly view', async () => {
      mockClient.get.mockResolvedValue({})
      await viewsApi.getWeeklyViewApi('1', '2026-01-06')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/views/weekly', { params: { weekStart: '2026-01-06' } })
    })

    it('getGanttViewApi should GET gantt view', async () => {
      mockClient.get.mockResolvedValue({})
      await viewsApi.getGanttViewApi('1', 'open')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/views/gantt', { params: { status: 'open' } })
    })

    it('getTableViewApi should GET table view with filter', async () => {
      mockClient.get.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 10 })
      await viewsApi.getTableViewApi('1', { priority: 'P0' })
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/views/table', { params: { priority: 'P0' } })
    })

    it('exportTableCsvApi should GET table export as blob', async () => {
      mockClient.get.mockResolvedValue(new Blob())
      await viewsApi.exportTableCsvApi('1')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/views/table/export', { params: undefined, responseType: 'blob' })
    })
  })

  describe('reports', () => {
    it('getWeeklyReportPreviewApi should GET preview', async () => {
      mockClient.get.mockResolvedValue({})
      await reportsApi.getWeeklyReportPreviewApi('1', '2026-01-06')
      expect(mockClient.get).toHaveBeenCalledWith('/teams/1/reports/weekly/preview', { params: { weekStart: '2026-01-06' } })
    })

    it('exportWeeklyReportApi should GET export as blob', async () => {
      mockClient.get.mockResolvedValue(new Blob())
      await reportsApi.exportWeeklyReportApi('1', '2026-01-06')
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
