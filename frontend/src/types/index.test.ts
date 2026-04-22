import { describe, it, expect } from 'vitest'
import type {
  User,
  Team,
  MainItem,
  SubItem,
  ProgressRecord,
  ItemPool,
  PermissionData,
  Role,
  PermissionGroup,
  PermissionItem,
} from '@/types'

describe('shared TypeScript interfaces', () => {
  it('should define a valid User', () => {
    const user: User = {
      id: 1,
      username: 'testuser',
      displayName: 'Test User',
      isSuperAdmin: false,
    }
    expect(user.username).toBe('testuser')
    expect(user.isSuperAdmin).toBe(false)
  })

  it('should define a valid Team', () => {
    const team: Team = {
      id: 1,
      name: 'Team Alpha',
      description: 'A team',
      pmId: 1,
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01',
    }
    expect(team.name).toBe('Team Alpha')
  })

  it('should define a valid MainItem', () => {
    const item: MainItem = {
      id: 1,
      teamId: 1,
      code: 'A001',
      title: 'Feature A',
      priority: 'P0',
      proposerId: 1,
      assigneeId: null,
      startDate: null,
      expectedEndDate: null,
      actualEndDate: null,
      status: 'pending',
      completion: 0,
      isKeyItem: false,
      delayCount: 0,
      archivedAt: null,
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01',
    }
    expect(item.code).toBe('A001')
  })

  it('should define a valid SubItem', () => {
    const sub: SubItem = {
      id: 1,
      teamId: 1,
      mainItemId: 1,
      code: 'A001-01',
      title: 'Sub task',
      description: 'Details',
      priority: 'P1',
      assigneeId: 2,
      startDate: '2026-01-01T00:00:00Z',
      expectedEndDate: '2026-01-15T00:00:00Z',
      actualEndDate: null,
      status: 'progressing',
      completion: 50,
      isKeyItem: true,
      delayCount: 0,
      weight: 1.5,
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01',
    }
    expect(sub.weight).toBe(1.5)
  })

  it('should define a valid ProgressRecord', () => {
    const record: ProgressRecord = {
      id: 1,
      subItemId: 1,
      teamId: 1,
      authorId: 1,
      completion: 60,
      achievement: 'Done something',
      blocker: '',
      lesson: '',
      isPMCorrect: false,
      createdAt: '2024-01-01',
    }
    expect(record.completion).toBe(60)
  })

  it('should define a valid ItemPool', () => {
    const pool: ItemPool = {
      id: 1,
      teamId: 1,
      title: 'Pool item',
      background: 'Context',
      expectedOutput: 'Result',
      submitterId: 1,
      status: '待分配',
      assignedMainId: null,
      assignedSubId: null,
      assignedMainCode: '',
      assignedMainTitle: '',
      assigneeId: null,
      rejectReason: '',
      reviewedAt: null,
      reviewerId: null,
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01',
    }
    expect(pool.status).toBe('待分配')
  })

  it('should define a valid PermissionData', () => {
    const permData: PermissionData = {
      isSuperAdmin: false,
      teamPermissions: { 1: ['team:read', 'team:write'] },
    }
    expect(permData.isSuperAdmin).toBe(false)
    expect(permData.teamPermissions[1]).toContain('team:read')
  })

  it('should define a valid Role', () => {
    const role: Role = {
      id: 1,
      name: 'PM',
      description: 'Project Manager',
      isPreset: true,
      permissionCount: 5,
      memberCount: 3,
      createdAt: '2024-01-01',
    }
    expect(role.name).toBe('PM')
    expect(role.isPreset).toBe(true)
  })

  it('should define a valid PermissionGroup', () => {
    const group: PermissionGroup = {
      resource: 'team',
      actions: [
        { code: 'team:read', description: 'View team' },
        { code: 'team:write', description: 'Edit team' },
      ],
    }
    expect(group.actions).toHaveLength(2)
  })

  it('should define a valid PermissionItem', () => {
    const item: PermissionItem = {
      code: 'item:create',
      description: 'Create items',
    }
    expect(item.code).toBe('item:create')
  })
})
