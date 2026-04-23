import { describe, it, expect } from 'vitest'
import { formatDate } from './format'

describe('formatDate', () => {
  it('returns "-" for null', () => {
    expect(formatDate(null)).toBe('-')
  })

  it('returns "-" for undefined', () => {
    expect(formatDate(undefined)).toBe('-')
  })

  it('returns "-" for empty string', () => {
    expect(formatDate('')).toBe('-')
  })

  it('replaces dashes with slashes', () => {
    expect(formatDate('2024-01-15')).toBe('2024/01/15')
  })

  it('handles date without dashes', () => {
    expect(formatDate('20240115')).toBe('20240115')
  })

  it('handles ISO datetime string', () => {
    expect(formatDate('2024-01-15T10:30:00')).toBe('2024/01/15T10:30:00')
  })
})
