import { useCallback } from 'react'

interface Member {
  userId: number
  displayName: string
}

/**
 * Pure function to look up a member's display name by assigneeId.
 * - Returns "Unassigned" for null assigneeId
 * - Returns "Unknown" for assigneeId not found in members
 */
export function getMemberName(members: Member[] | undefined, assigneeId: number | null): string {
  if (assigneeId === null) return 'Unassigned'
  if (!members) return 'Unknown'
  const m = members.find((m) => m.userId === assigneeId)
  return m ? m.displayName : 'Unknown'
}

/**
 * Hook that returns a memoized function to look up member names.
 */
export function useMemberName(members: Member[] | undefined) {
  return useCallback(
    (assigneeId: number | null): string => getMemberName(members, assigneeId),
    [members],
  )
}
