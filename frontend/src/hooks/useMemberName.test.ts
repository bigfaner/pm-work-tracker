import { describe, it, expect } from 'vitest'
import { getMemberName } from './useMemberName'

const members = [
  { userId: 1, displayName: 'Alice', username: 'alice', role: 'pm', joinedAt: '2024-01-01' },
  { userId: 2, displayName: 'Bob', username: 'bob', role: 'member', joinedAt: '2024-01-01' },
]

describe('getMemberName', () => {
  it('returns "Unassigned" for null assigneeId', () => {
    expect(getMemberName(members, null)).toBe('Unassigned')
  })

  it('returns "Unknown" for assigneeId not found in members', () => {
    expect(getMemberName(members, 999)).toBe('Unknown')
  })

  it('returns displayName for valid assigneeId', () => {
    expect(getMemberName(members, 1)).toBe('Alice')
  })

  it('returns "Unassigned" when members is undefined and assigneeId is null', () => {
    expect(getMemberName(undefined, null)).toBe('Unassigned')
  })

  it('returns "Unknown" for non-null id when members is undefined', () => {
    expect(getMemberName(undefined, 1)).toBe('Unknown')
  })

  it('returns "Unknown" for non-null id when members is empty', () => {
    expect(getMemberName([], 1)).toBe('Unknown')
  })
})
