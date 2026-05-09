/**
 * Format a date string by replacing dashes with slashes.
 * Returns "-" for null, undefined, or empty strings.
 */
export function formatDate(date: string | null | undefined): string {
  if (!date) return '-'
  return date.replace(/-/g, '/')
}

/**
 * Format an ISO datetime string as a date-only string.
 * Truncates to the first 10 characters (YYYY-MM-DD), then replaces dashes with slashes.
 * Returns "-" for null, undefined, or empty strings.
 */
export function formatDateOnly(dateStr?: string | null): string {
  if (!dateStr) return '-'
  return formatDate(dateStr.slice(0, 10))
}
