import { create } from 'zustand'
import type { Team } from '@/types'

interface TeamState {
  currentTeamId: number | null
  teams: Team[]
  setCurrentTeam: (teamId: number | null) => void
  setTeams: (teams: Team[]) => void
}

export const useTeamStore = create<TeamState>((set) => ({
  currentTeamId: null,
  teams: [],
  setCurrentTeam: (teamId) => set({ currentTeamId: teamId }),
  setTeams: (teams) => set({ teams }),
}))
