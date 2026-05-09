import { describe, it, expect } from 'vitest'
import {
  getCurrentWeekStart,
  getWeekNumber,
  getISOWeekYear,
  toLocalDateString,
  addWeeks,
  formatWeekLabel,
} from './weekUtils'

describe('toLocalDateString', () => {
  it('formats a Date to YYYY-MM-DD', () => {
    expect(toLocalDateString(new Date(2026, 3, 13))).toBe('2026-04-13')
  })

  it('pads month and day with zeros', () => {
    expect(toLocalDateString(new Date(2026, 0, 5))).toBe('2026-01-05')
  })
})

describe('getWeekNumber', () => {
  it('returns correct week number for a normal date', () => {
    expect(getWeekNumber('2026-04-13')).toBe(16)
  })

  it('returns week 1 for Jan 1 when it is in week 1', () => {
    // 2026-01-01 is Thursday, so it's in ISO week 1
    expect(getWeekNumber('2026-01-01')).toBe(1)
  })

  it('handles cross-year boundary: Dec 28 2026 is in week 53 of 2026', () => {
    // 2026-12-28 is Monday, ISO week 53 of 2026
    expect(getWeekNumber('2026-12-28')).toBe(53)
  })

  it('handles cross-year boundary: Dec 31 2018 is in week 1 of 2019', () => {
    // 2018-12-31 is Monday, ISO week 1 of 2019
    expect(getWeekNumber('2018-12-31')).toBe(1)
  })
})

describe('getISOWeekYear', () => {
  it('returns the natural year for a normal date', () => {
    expect(getISOWeekYear('2026-04-13')).toBe(2026)
  })

  it('returns next year for Dec 31 2018 (ISO week 1 of 2019)', () => {
    expect(getISOWeekYear('2018-12-31')).toBe(2019)
  })

  it('returns previous year for Jan 1 2016 (ISO week 53 of 2015)', () => {
    // 2016-01-01 is Friday, belongs to ISO week 53 of 2015
    expect(getISOWeekYear('2016-01-01')).toBe(2015)
  })
})

describe('getCurrentWeekStart', () => {
  it('returns a string in YYYY-MM-DD format', () => {
    const result = getCurrentWeekStart()
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })

  it('returns a Monday', () => {
    const result = getCurrentWeekStart()
    const [y, m, d] = result.split('-').map(Number)
    const date = new Date(y, m - 1, d)
    expect(date.getDay()).toBe(1) // 1 = Monday
  })
})

describe('addWeeks', () => {
  it('adds 1 week', () => {
    expect(addWeeks('2026-04-13', 1)).toBe('2026-04-20')
  })

  it('subtracts 1 week', () => {
    expect(addWeeks('2026-04-13', -1)).toBe('2026-04-06')
  })

  it('adds 0 weeks returns same date', () => {
    expect(addWeeks('2026-04-13', 0)).toBe('2026-04-13')
  })

  it('crosses month boundary correctly', () => {
    expect(addWeeks('2026-04-27', 1)).toBe('2026-05-04')
  })

  it('crosses year boundary correctly', () => {
    expect(addWeeks('2025-12-29', 1)).toBe('2026-01-05')
  })

  it('DST boundary: spring forward (US 2026-03-08)', () => {
    // 2026-03-02 + 1 week should be 2026-03-09 (crosses DST spring-forward on 2026-03-08)
    expect(addWeeks('2026-03-02', 1)).toBe('2026-03-09')
  })

  it('DST boundary: fall back (US 2026-11-01)', () => {
    // 2026-10-26 + 1 week should be 2026-11-02 (crosses DST fall-back on 2026-11-01)
    expect(addWeeks('2026-10-26', 1)).toBe('2026-11-02')
  })
})

describe('formatWeekLabel', () => {
  it('formats week label correctly', () => {
    expect(formatWeekLabel('2026-04-13')).toBe('2026年第16周  04/13 ~ 04/19')
  })

  it('pads month and day with zeros', () => {
    expect(formatWeekLabel('2026-01-05')).toBe('2026年第2周  01/05 ~ 01/11')
  })

  it('handles cross-year week: 2018-12-31 is week 1 of 2019', () => {
    expect(formatWeekLabel('2018-12-31')).toBe('2019年第1周  12/31 ~ 01/06')
  })
})
