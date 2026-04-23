import { describe, it, expect } from 'vitest'
import { MAIN_ITEM_STATUSES, SUB_ITEM_STATUSES, isOverdue, STATUS_OPTIONS, MAIN_TERMINAL_STATUSES, SUB_TERMINAL_STATUSES } from './status'

describe('MAIN_ITEM_STATUSES', () => {
  it('has 7 entries', () => {
    expect(Object.keys(MAIN_ITEM_STATUSES)).toHaveLength(7)
  })

  it('contains all required codes', () => {
    const codes = Object.keys(MAIN_ITEM_STATUSES)
    expect(codes).toContain('pending')
    expect(codes).toContain('progressing')
    expect(codes).toContain('blocking')
    expect(codes).toContain('pausing')
    expect(codes).toContain('reviewing')
    expect(codes).toContain('completed')
    expect(codes).toContain('closed')
  })

  it('each entry has name, variant, terminal fields', () => {
    for (const [, def] of Object.entries(MAIN_ITEM_STATUSES)) {
      expect(def).toHaveProperty('name')
      expect(def).toHaveProperty('variant')
      expect(def).toHaveProperty('terminal')
      expect(typeof def.name).toBe('string')
      expect(typeof def.variant).toBe('string')
      expect(typeof def.terminal).toBe('boolean')
    }
  })
})

describe('SUB_ITEM_STATUSES', () => {
  it('has 6 entries', () => {
    expect(Object.keys(SUB_ITEM_STATUSES)).toHaveLength(6)
  })

  it('contains all required codes (no reviewing)', () => {
    const codes = Object.keys(SUB_ITEM_STATUSES)
    expect(codes).toContain('pending')
    expect(codes).toContain('progressing')
    expect(codes).toContain('blocking')
    expect(codes).toContain('pausing')
    expect(codes).toContain('completed')
    expect(codes).toContain('closed')
    expect(codes).not.toContain('reviewing')
  })

  it('each entry has name, variant, terminal fields', () => {
    for (const [, def] of Object.entries(SUB_ITEM_STATUSES)) {
      expect(def).toHaveProperty('name')
      expect(def).toHaveProperty('variant')
      expect(def).toHaveProperty('terminal')
      expect(typeof def.name).toBe('string')
      expect(typeof def.variant).toBe('string')
      expect(typeof def.terminal).toBe('boolean')
    }
  })
})

describe('STATUS_OPTIONS', () => {
  it('contains main item status codes for filter dropdowns', () => {
    expect(STATUS_OPTIONS).toEqual(
      expect.arrayContaining(['pending', 'progressing', 'blocking', 'pausing', 'reviewing', 'completed', 'closed']),
    )
  })
})

describe('isOverdue', () => {
  it('returns false when expectedEndDate is empty', () => {
    expect(isOverdue('', 'progressing')).toBe(false)
  })

  it('returns false when expectedEndDate is undefined', () => {
    expect(isOverdue(undefined, 'progressing')).toBe(false)
  })

  it('returns false when status is undefined', () => {
    expect(isOverdue('2020-01-01', undefined)).toBe(false)
  })

  it('returns false for terminal status completed', () => {
    expect(isOverdue('2020-01-01', 'completed')).toBe(false)
  })

  it('returns false for terminal status closed', () => {
    expect(isOverdue('2020-01-01', 'closed')).toBe(false)
  })

  it('returns true when date is past and status is non-terminal', () => {
    expect(isOverdue('2020-01-01', 'progressing')).toBe(true)
  })

  it('returns false when date is in the future and status is non-terminal', () => {
    const future = new Date()
    future.setFullYear(future.getFullYear() + 1)
    expect(isOverdue(future.toISOString().slice(0, 10), 'progressing')).toBe(false)
  })

  it('returns true for sub-item status blocking with past date', () => {
    expect(isOverdue('2020-01-01', 'blocking')).toBe(true)
  })
})

describe('MAIN_TERMINAL_STATUSES', () => {
  it('contains completed and closed', () => {
    expect(MAIN_TERMINAL_STATUSES).toContain('completed')
    expect(MAIN_TERMINAL_STATUSES).toContain('closed')
  })

  it('has exactly 2 entries', () => {
    expect(MAIN_TERMINAL_STATUSES).toHaveLength(2)
  })
})

describe('SUB_TERMINAL_STATUSES', () => {
  it('contains completed and closed', () => {
    expect(SUB_TERMINAL_STATUSES).toContain('completed')
    expect(SUB_TERMINAL_STATUSES).toContain('closed')
  })

  it('has exactly 2 entries', () => {
    expect(SUB_TERMINAL_STATUSES).toHaveLength(2)
  })
})
