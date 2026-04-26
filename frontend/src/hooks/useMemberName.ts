import { useCallback } from 'react'

interface Member {
  bizKey: string
  displayName: string
}

export function getMemberName(members: Member[] | undefined, assigneeKey: string | null): string {
  if (assigneeKey === null) return 'Unassigned'
  if (!members) return 'Unknown'
  const m = members.find((m) => m.bizKey === assigneeKey)
  return m ? m.displayName : 'Unknown'
}

export function useMemberName(members: Member[] | undefined) {
  return useCallback(
    (assigneeKey: string | null): string => getMemberName(members, assigneeKey),
    [members],
  )
}
