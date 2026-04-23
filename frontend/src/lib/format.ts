/**
 * Format a date string by replacing dashes with slashes.
 * Returns "-" for null, undefined, or empty strings.
 */
export function formatDate(date: string | null | undefined): string {
  if (!date) return '-'
  return date.replace(/-/g, '/')
}
