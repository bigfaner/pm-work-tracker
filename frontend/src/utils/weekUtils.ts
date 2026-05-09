/** Formats a Date to YYYY-MM-DD using local timezone */
export function toLocalDateString(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/** Returns the ISO week number (1-53) for a given YYYY-MM-DD date string */
export function getWeekNumber(dateStr: string): number {
  const [y, m, d] = dateStr.split('-').map(Number)
  const date = new Date(y, m - 1, d)
  const dayNum = date.getDay() || 7
  const thursday = new Date(date)
  thursday.setDate(date.getDate() + 4 - dayNum)
  const jan1 = new Date(thursday.getFullYear(), 0, 1)
  const dayOfYear = Math.floor((thursday.getTime() - jan1.getTime()) / 86400000) + 1
  return Math.ceil(dayOfYear / 7)
}

/** Returns the ISO week year for a given YYYY-MM-DD date string.
 *  May differ from the calendar year near year boundaries (ISO 8601). */
export function getISOWeekYear(dateStr: string): number {
  const [y, m, d] = dateStr.split('-').map(Number)
  const date = new Date(y, m - 1, d)
  const dayNum = date.getDay() || 7
  const thursday = new Date(date)
  thursday.setDate(date.getDate() + 4 - dayNum)
  return thursday.getFullYear()
}

/** Returns the Monday of the current week as YYYY-MM-DD */
export function getCurrentWeekStart(): string {
  const now = new Date()
  const day = now.getDay()
  const diff = now.getDate() - day + (day === 0 ? -6 : 1)
  const monday = new Date(now)
  monday.setDate(diff)
  return toLocalDateString(monday)
}

/** Adds n weeks to weekStart (YYYY-MM-DD), returns new YYYY-MM-DD.
 *  Uses setDate to avoid DST ±1 hour issues. */
export function addWeeks(weekStart: string, n: number): string {
  const [y, m, d] = weekStart.split('-').map(Number)
  const date = new Date(y, m - 1, d)
  date.setDate(date.getDate() + n * 7)
  return toLocalDateString(date)
}

/** Formats weekStart as "2026年第16周  04/13 ~ 04/19" */
export function formatWeekLabel(weekStart: string): string {
  const isoYear = getISOWeekYear(weekStart)
  const weekNum = getWeekNumber(weekStart)
  const [y, m, d] = weekStart.split('-').map(Number)
  const start = new Date(y, m - 1, d)
  const end = new Date(start)
  end.setDate(start.getDate() + 6)
  const sm = String(start.getMonth() + 1).padStart(2, '0')
  const sd = String(start.getDate()).padStart(2, '0')
  const em = String(end.getMonth() + 1).padStart(2, '0')
  const ed = String(end.getDate()).padStart(2, '0')
  return `${isoYear}年第${weekNum}周  ${sm}/${sd} ~ ${em}/${ed}`
}
