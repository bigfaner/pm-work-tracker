import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Team } from '@/types'

interface TeamState {
  currentTeamId: number | null
  teams: Team[]
  setCurrentTeam: (teamId: number | null) => void
  setTeams: (teams: Team[]) => void
}

export const useTeamStore = create<TeamState>()(
  persist(
    (set) => ({
      currentTeamId: null,
      teams: [],
      setCurrentTeam: (teamId) => set({ currentTeamId: teamId }),
      setTeams: (teams) => set({ teams }),
    }),
    { name: 'team-storage' },
  ),
)
