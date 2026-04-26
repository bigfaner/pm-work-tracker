import { describe, it, expect, beforeEach } from 'vitest'
import { useTeamStore } from './team'
import type { Team } from '@/types'

const mockTeams: Team[] = [
  {
    bizKey: 'team-1',
    name: 'Team Alpha',
    description: 'First team',
    pmKey: 'pm-1',
    createdAt: '2024-01-01',
    updatedAt: '2024-01-01',
  },
  {
    bizKey: 'team-2',
    name: 'Team Beta',
    description: 'Second team',
    pmKey: 'pm-2',
    createdAt: '2024-01-01',
    updatedAt: '2024-01-01',
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
    useTeamStore.getState().setCurrentTeam('team-1')
    expect(useTeamStore.getState().currentTeamId).toBe('team-1')
  })

  it('setCurrentTeam can be set back to null', () => {
    useTeamStore.getState().setCurrentTeam('team-1')
    useTeamStore.getState().setCurrentTeam(null)
    expect(useTeamStore.getState().currentTeamId).toBeNull()
  })
})
