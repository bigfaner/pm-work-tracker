import { describe, it, expect, beforeEach } from 'vitest'
import { useTeamStore } from './team'
import type { Team } from '@/types'

const mockTeams: Team[] = [
  {
    id: 1,
    name: 'Team Alpha',
    description: 'First team',
    pm_id: 1,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    name: 'Team Beta',
    description: 'Second team',
    pm_id: 2,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
]

describe('useTeamStore', () => {
  beforeEach(() => {
    useTeamStore.getState().setCurrentTeam(null)
    useTeamStore.getState().setTeams([])
  })

  it('starts with null currentTeamId and empty teams', () => {
    const state = useTeamStore.getState()
    expect(state.currentTeamId).toBeNull()
    expect(state.teams).toEqual([])
  })

  it('setTeams populates teams list', () => {
    useTeamStore.getState().setTeams(mockTeams)
    expect(useTeamStore.getState().teams).toEqual(mockTeams)
  })

  it('setCurrentTeam sets currentTeamId', () => {
    useTeamStore.getState().setCurrentTeam(1)
    expect(useTeamStore.getState().currentTeamId).toBe(1)
  })

  it('setCurrentTeam can be set back to null', () => {
    useTeamStore.getState().setCurrentTeam(1)
    useTeamStore.getState().setCurrentTeam(null)
    expect(useTeamStore.getState().currentTeamId).toBeNull()
  })
})
