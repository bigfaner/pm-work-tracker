import { describe, it, expect, vi, beforeEach } from 'vitest'
import client from './client'
import {
  createMilestoneMapApi,
  listMilestoneMapsApi,
  getMilestoneMapApi,
  updateMilestoneMapApi,
  deleteMilestoneMapApi,
  changeMilestoneMapStatusApi,
  getMilestoneMapTransitionsApi,
  createMilestoneApi,
  listMilestonesByMapApi,
  listMilestonesByTeamApi,
  getMilestoneApi,
  updateMilestoneApi,
  deleteMilestoneApi,
  changeMilestoneStatusApi,
  getMilestoneTransitionsApi,
} from './milestones'
import type { MilestoneMap, Milestone } from '@/types'

vi.mock('./client', () => ({
  default: { post: vi.fn(), get: vi.fn(), put: vi.fn(), delete: vi.fn() },
}))

const mockClient = client as unknown as {
  post: ReturnType<typeof vi.fn>
  get: ReturnType<typeof vi.fn>
  put: ReturnType<typeof vi.fn>
  delete: ReturnType<typeof vi.fn>
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe('MilestoneMap API', () => {
  const teamBizKey = 'team-123'
  const mapBizKey = 'map-456'
  const mockMap: MilestoneMap = {
    bizKey: mapBizKey,
    teamKey: teamBizKey,
    mapName: 'Q3 Plan',
    mapDesc: 'Q3 milestone plan',
    mapStatus: 'planning',
    statusName: '规划中',
    milestoneCount: 0,
    itemCount: 0,
    overallProgress: 0,
    createTime: '2026-05-01T00:00:00Z',
    dbUpdateTime: '2026-05-01T00:00:00Z',
  }

  describe('createMilestoneMapApi', () => {
    it('should POST to /teams/:teamBizKey/milestone-maps', async () => {
      mockClient.post.mockResolvedValue(mockMap)

      const result = await createMilestoneMapApi(teamBizKey, {
        mapName: 'Q3 Plan',
        mapDesc: 'Q3 milestone plan',
      })

      expect(mockClient.post).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps`,
        { mapName: 'Q3 Plan', mapDesc: 'Q3 milestone plan' },
      )
      expect(result).toEqual(mockMap)
    })
  })

  describe('listMilestoneMapsApi', () => {
    it('should GET /teams/:teamBizKey/milestone-maps with filter params', async () => {
      const pageResult = { items: [mockMap], total: 1, page: 1, size: 20 }
      mockClient.get.mockResolvedValue(pageResult)

      const result = await listMilestoneMapsApi(teamBizKey, {
        status: 'planning',
        page: 1,
        pageSize: 20,
      })

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps`,
        { params: { status: 'planning', page: 1, pageSize: 20 } },
      )
      expect(result).toEqual(pageResult)
    })

    it('should GET without filter params when not provided', async () => {
      const pageResult = { items: [], total: 0, page: 1, size: 20 }
      mockClient.get.mockResolvedValue(pageResult)

      const result = await listMilestoneMapsApi(teamBizKey)

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps`,
        { params: undefined },
      )
      expect(result).toEqual(pageResult)
    })
  })

  describe('getMilestoneMapApi', () => {
    it('should GET /teams/:teamBizKey/milestone-maps/:mapId', async () => {
      mockClient.get.mockResolvedValue(mockMap)

      const result = await getMilestoneMapApi(teamBizKey, mapBizKey)

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps/${mapBizKey}`,
      )
      expect(result).toEqual(mockMap)
    })
  })

  describe('updateMilestoneMapApi', () => {
    it('should PUT to /teams/:teamBizKey/milestone-maps/:mapId', async () => {
      const updated = { ...mockMap, mapName: 'Q3 Plan v2' }
      mockClient.put.mockResolvedValue(updated)

      const result = await updateMilestoneMapApi(teamBizKey, mapBizKey, {
        mapName: 'Q3 Plan v2',
      })

      expect(mockClient.put).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps/${mapBizKey}`,
        { mapName: 'Q3 Plan v2' },
      )
      expect(result).toEqual(updated)
    })
  })

  describe('deleteMilestoneMapApi', () => {
    it('should DELETE /teams/:teamBizKey/milestone-maps/:mapId', async () => {
      mockClient.delete.mockResolvedValue({ message: 'deleted' })

      await deleteMilestoneMapApi(teamBizKey, mapBizKey)

      expect(mockClient.delete).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps/${mapBizKey}`,
      )
    })
  })

  describe('changeMilestoneMapStatusApi', () => {
    it('should PUT to /teams/:teamBizKey/milestone-maps/:mapId/status', async () => {
      const updated = { ...mockMap, mapStatus: 'reviewed', statusName: '已评审' }
      mockClient.put.mockResolvedValue(updated)

      const result = await changeMilestoneMapStatusApi(teamBizKey, mapBizKey, {
        status: 'reviewed',
      })

      expect(mockClient.put).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps/${mapBizKey}/status`,
        { status: 'reviewed' },
      )
      expect(result).toEqual(updated)
    })
  })

  describe('getMilestoneMapTransitionsApi', () => {
    it('should GET available-transitions and unwrap transitions array', async () => {
      mockClient.get.mockResolvedValue({ transitions: ['reviewed', 'ready'] })

      const result = await getMilestoneMapTransitionsApi(teamBizKey, mapBizKey)

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps/${mapBizKey}/available-transitions`,
      )
      expect(result).toEqual(['reviewed', 'ready'])
    })

    it('should return empty array when transitions is null', async () => {
      mockClient.get.mockResolvedValue({ transitions: null })

      const result = await getMilestoneMapTransitionsApi(teamBizKey, mapBizKey)

      expect(result).toEqual([])
    })
  })
})

describe('Milestone API', () => {
  const teamBizKey = 'team-123'
  const mapBizKey = 'map-456'
  const milestoneBizKey = 'ms-789'
  const mockMilestone: Milestone = {
    bizKey: milestoneBizKey,
    teamKey: teamBizKey,
    milestoneMapKey: mapBizKey,
    milestoneName: 'M1',
    expectedEndDate: '2026-06-30',
    milestoneStatus: 'not_started',
    statusName: '未开始',
    completion: 0,
    relatedMICount: 0,
    createTime: '2026-05-01T00:00:00Z',
    dbUpdateTime: '2026-05-01T00:00:00Z',
  }

  describe('createMilestoneApi', () => {
    it('should POST to /teams/:teamBizKey/milestone-maps/:mapId/milestones', async () => {
      mockClient.post.mockResolvedValue(mockMilestone)

      const result = await createMilestoneApi(teamBizKey, mapBizKey, {
        milestoneName: 'M1',
        expectedEndDate: '2026-06-30',
      })

      expect(mockClient.post).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`,
        { milestoneName: 'M1', expectedEndDate: '2026-06-30' },
      )
      expect(result).toEqual(mockMilestone)
    })
  })

  describe('listMilestonesByMapApi', () => {
    it('should GET /teams/:teamBizKey/milestone-maps/:mapId/milestones', async () => {
      const listResult = { items: [mockMilestone], total: 1 }
      mockClient.get.mockResolvedValue(listResult)

      const result = await listMilestonesByMapApi(teamBizKey, mapBizKey)

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`,
      )
      expect(result).toEqual(listResult)
    })
  })

  describe('listMilestonesByTeamApi', () => {
    it('should GET /teams/:teamBizKey/milestones with excludeCancelled param', async () => {
      const listResult = { items: [mockMilestone], total: 1 }
      mockClient.get.mockResolvedValue(listResult)

      const result = await listMilestonesByTeamApi(teamBizKey, { excludeCancelled: true })

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestones`,
        { params: { excludeCancelled: true } },
      )
      expect(result).toEqual(listResult)
    })

    it('should GET without params when not provided', async () => {
      const listResult = { items: [mockMilestone], total: 1 }
      mockClient.get.mockResolvedValue(listResult)

      const result = await listMilestonesByTeamApi(teamBizKey)

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestones`,
        { params: undefined },
      )
      expect(result).toEqual(listResult)
    })
  })

  describe('getMilestoneApi', () => {
    it('should GET /teams/:teamBizKey/milestones/:milestoneId', async () => {
      mockClient.get.mockResolvedValue(mockMilestone)

      const result = await getMilestoneApi(teamBizKey, milestoneBizKey)

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestones/${milestoneBizKey}`,
      )
      expect(result).toEqual(mockMilestone)
    })
  })

  describe('updateMilestoneApi', () => {
    it('should PUT to /teams/:teamBizKey/milestones/:milestoneId', async () => {
      const updated = { ...mockMilestone, milestoneName: 'M1 v2' }
      mockClient.put.mockResolvedValue(updated)

      const result = await updateMilestoneApi(teamBizKey, milestoneBizKey, {
        milestoneName: 'M1 v2',
      })

      expect(mockClient.put).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestones/${milestoneBizKey}`,
        { milestoneName: 'M1 v2' },
      )
      expect(result).toEqual(updated)
    })
  })

  describe('deleteMilestoneApi', () => {
    it('should DELETE /teams/:teamBizKey/milestones/:milestoneId', async () => {
      mockClient.delete.mockResolvedValue({ message: 'deleted' })

      await deleteMilestoneApi(teamBizKey, milestoneBizKey)

      expect(mockClient.delete).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestones/${milestoneBizKey}`,
      )
    })
  })

  describe('changeMilestoneStatusApi', () => {
    it('should PUT to /teams/:teamBizKey/milestones/:milestoneId/status', async () => {
      const updated = { ...mockMilestone, milestoneStatus: 'in_progress', statusName: '进行中' }
      mockClient.put.mockResolvedValue(updated)

      const result = await changeMilestoneStatusApi(teamBizKey, milestoneBizKey, {
        status: 'in_progress',
      })

      expect(mockClient.put).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestones/${milestoneBizKey}/status`,
        { status: 'in_progress' },
      )
      expect(result).toEqual(updated)
    })
  })

  describe('getMilestoneTransitionsApi', () => {
    it('should GET available-transitions and unwrap transitions array', async () => {
      mockClient.get.mockResolvedValue({ transitions: ['in_progress', 'cancelled'] })

      const result = await getMilestoneTransitionsApi(teamBizKey, milestoneBizKey)

      expect(mockClient.get).toHaveBeenCalledWith(
        `/teams/${teamBizKey}/milestones/${milestoneBizKey}/available-transitions`,
      )
      expect(result).toEqual(['in_progress', 'cancelled'])
    })

    it('should return empty array when transitions is null', async () => {
      mockClient.get.mockResolvedValue({ transitions: null })

      const result = await getMilestoneTransitionsApi(teamBizKey, milestoneBizKey)

      expect(result).toEqual([])
    })
  })
})
