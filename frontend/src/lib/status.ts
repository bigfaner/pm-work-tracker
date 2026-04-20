export const MAIN_ITEM_STATUSES = {
  pending:     { name: '待开始', variant: 'planning',    terminal: false },
  progressing: { name: '进行中', variant: 'in-progress', terminal: false },
  blocking:    { name: '阻塞中', variant: 'overdue',     terminal: false },
  pausing:     { name: '已暂停', variant: 'on-hold',     terminal: false },
  reviewing:   { name: '待验收', variant: 'pending',     terminal: false },
  completed:   { name: '已完成', variant: 'completed',   terminal: true },
  closed:      { name: '已关闭', variant: 'cancelled',   terminal: true },
} as const

export const SUB_ITEM_STATUSES = {
  pending:     { name: '待开始', variant: 'planning',    terminal: false },
  progressing: { name: '进行中', variant: 'in-progress', terminal: false },
  blocking:    { name: '阻塞中', variant: 'overdue',     terminal: false },
  pausing:     { name: '已暂停', variant: 'on-hold',     terminal: false },
  completed:   { name: '已完成', variant: 'completed',   terminal: true },
  closed:      { name: '已关闭', variant: 'cancelled',   terminal: true },
} as const

/** English status codes for filter dropdowns (all main item statuses) */
export const STATUS_OPTIONS = Object.keys(MAIN_ITEM_STATUSES)

/** Lookup variant from either main or sub item status maps */
export function getStatusVariant(status: string): string {
  const def = (MAIN_ITEM_STATUSES as Record<string, { variant: string }>)[status]
    || (SUB_ITEM_STATUSES as Record<string, { variant: string }>)[status]
  return def?.variant ?? 'default'
}

/** Lookup Chinese display name from either main or sub item status maps */
export function getStatusName(status: string): string | undefined {
  const def = (MAIN_ITEM_STATUSES as Record<string, { name: string }>)[status]
    || (SUB_ITEM_STATUSES as Record<string, { name: string }>)[status]
  return def?.name
}

/** Check if an item is overdue based on expected end date and status code */
export function isOverdue(expectedEndDate?: string, status?: string): boolean {
  if (!expectedEndDate || !status) return false
  const isTerminal = (MAIN_ITEM_STATUSES as Record<string, { terminal: boolean }>)[status]?.terminal
    || (SUB_ITEM_STATUSES as Record<string, { terminal: boolean }>)[status]?.terminal
  if (isTerminal) return false
  return new Date(expectedEndDate) < new Date()
}
