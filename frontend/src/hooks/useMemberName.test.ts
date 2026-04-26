import { describe, it, expect } from 'vitest'
import { getMemberName } from './useMemberName'

const members = [
  { bizKey: 'U001', displayName: 'Alice', username: 'alice', role: 'pm', joinedAt: '2024-01-01' },
  { bizKey: 'U002', displayName: 'Bob', username: 'bob', role: 'member', joinedAt: '2024-01-01' },
]

describe('getMemberName', () => {
  it('returns "Unassigned" for null assigneeKey', () => {
    expect(getMemberName(members, null)).toBe('Unassigned')
  })

  it('returns "Unknown" for assigneeKey not found in members', () => {
    expect(getMemberName(members, 'U999')).toBe('Unknown')
  })

  it('returns displayName for valid assigneeKey', () => {
    expect(getMemberName(members, 'U001')).toBe('Alice')
  })

  it('returns "Unassigned" when members is undefined and assigneeKey is null', () => {
    expect(getMemberName(undefined, null)).toBe('Unassigned')
  })

  it('returns "Unknown" for non-null key when members is undefined', () => {
    expect(getMemberName(undefined, 'U001')).toBe('Unknown')
  })

  it('returns "Unknown" for non-null key when members is empty', () => {
    expect(getMemberName([], 'U001')).toBe('Unknown')
  })
})
