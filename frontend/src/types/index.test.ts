import { describe, it, expect } from 'vitest'
import type {
  User,
  Team,
  TeamMember,
  MainItem,
  SubItem,
  ProgressRecord,
  ItemPool,
} from '@/types'

describe('shared TypeScript interfaces', () => {
  it('should define a valid User', () => {
    const user: User = {
      id: 1,
      username: 'testuser',
      display_name: 'Test User',
      is_super_admin: false,
      can_create_team: false,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    }
    expect(user.username).toBe('testuser')
    expect(user.is_super_admin).toBe(false)
  })

  it('should define a valid Team', () => {
    const team: Team = {
      id: 1,
      name: 'Team Alpha',
      description: 'A team',
      pm_id: 1,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    }
    expect(team.name).toBe('Team Alpha')
  })

  it('should define a valid TeamMember', () => {
    const member: TeamMember = {
      id: 1,
      team_id: 1,
      user_id: 1,
      role: 'member',
      joined_at: '2026-01-01T00:00:00Z',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    }
    expect(member.role).toBe('member')
  })

  it('should define a valid MainItem', () => {
    const item: MainItem = {
      id: 1,
      team_id: 1,
      code: 'A001',
      title: 'Feature A',
      priority: 'P0',
      proposer_id: 1,
      assignee_id: null,
      start_date: null,
      expected_end_date: null,
      actual_end_date: null,
      status: '待开始',
      completion: 0,
      is_key_item: false,
      delay_count: 0,
      archived_at: null,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    }
    expect(item.code).toBe('A001')
  })

  it('should define a valid SubItem', () => {
    const sub: SubItem = {
      id: 1,
      team_id: 1,
      main_item_id: 1,
      title: 'Sub task',
      description: 'Details',
      priority: 'P1',
      assignee_id: 2,
      start_date: '2026-01-01T00:00:00Z',
      expected_end_date: '2026-01-15T00:00:00Z',
      actual_end_date: null,
      status: '进行中',
      completion: 50,
      is_key_item: true,
      delay_count: 0,
      weight: 1.5,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    }
    expect(sub.weight).toBe(1.5)
  })

  it('should define a valid ProgressRecord', () => {
    const record: ProgressRecord = {
      id: 1,
      sub_item_id: 1,
      team_id: 1,
      author_id: 1,
      completion: 60,
      achievement: 'Done something',
      blocker: '',
      lesson: '',
      is_pm_correct: false,
      created_at: '2026-01-01T00:00:00Z',
    }
    expect(record.completion).toBe(60)
  })

  it('should define a valid ItemPool', () => {
    const pool: ItemPool = {
      id: 1,
      team_id: 1,
      title: 'Pool item',
      background: 'Context',
      expected_output: 'Result',
      submitter_id: 1,
      status: '待分配',
      assigned_main_id: null,
      assigned_sub_id: null,
      assignee_id: null,
      reject_reason: '',
      reviewed_at: null,
      reviewer_id: null,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    }
    expect(pool.status).toBe('待分配')
  })
})
