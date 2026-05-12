/**
 * Build a full URL for sharing (clipboard, external links).
 * Includes VITE_BASE_PATH so links work when deployed under a sub-path.
 */
export function buildFullUrl(
  path: string,
  basePath: string = import.meta.env.VITE_BASE_PATH ?? '',
): string {
  return `${window.location.origin}${basePath}${path}`
}
