/**
 * Standalone toast shim for use in non-React modules (e.g., Axios interceptors).
 *
 * ToastProvider calls `init(showToastFn)` on mount to wire up the real toast
 * implementation. Until then, calls are silently ignored.
 */

type ToastVariant = "default" | "success" | "error" | "warning";

type ShowToastFn = (message: string, variant: ToastVariant) => void;

let showToastFn: ShowToastFn | null = null;

export function init(fn: ShowToastFn): void {
  showToastFn = fn;
}

export function showToast(
  message: string,
  variant: ToastVariant = "default",
): void {
  showToastFn?.(message, variant);
}
