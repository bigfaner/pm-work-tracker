import * as React from 'react'
import { cn } from '@/lib/utils'

interface ToastProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'success' | 'error' | 'warning'
}

const toastVariants = {
  default: 'bg-white border-border text-primary',
  success: 'bg-success-bg border-success-text/20 text-success-text',
  error: 'bg-error-bg border-error-text/20 text-error-text',
  warning: 'bg-warning-bg border-warning-text/20 text-warning-text',
}

function Toast({ variant = 'default', className, ...props }: ToastProps) {
  return (
    <div
      role="alert"
      className={cn(
        'flex items-center gap-2 rounded-lg px-4 py-3 text-[13px] font-medium border shadow-md',
        toastVariants[variant],
        className
      )}
      {...props}
    />
  )
}

interface ToastProviderProps {
  children: React.ReactNode
}

interface ToastItem {
  id: string
  message: string
  variant: 'default' | 'success' | 'error' | 'warning'
}

interface ToastContextValue {
  toasts: ToastItem[]
  addToast: (message: string, variant?: ToastItem['variant']) => void
  removeToast: (id: string) => void
}

const ToastContext = React.createContext<ToastContextValue | null>(null)

function ToastProvider({ children }: ToastProviderProps) {
  const [toasts, setToasts] = React.useState<ToastItem[]>([])

  const addToast = React.useCallback(
    (message: string, variant: ToastItem['variant'] = 'default') => {
      const id = Math.random().toString(36).slice(2, 9)
      setToasts((prev) => [...prev, { id, message, variant }])
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id))
      }, 1000)
    },
    []
  )

  const removeToast = React.useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
      {children}
      <div className="fixed top-4 right-4 z-[200] flex flex-col gap-2">
        {toasts.map((toast) => (
          <Toast
            key={toast.id}
            variant={toast.variant}
            onClick={() => removeToast(toast.id)}
            className="cursor-pointer animate-[slideIn_0.2s_ease-out]"
          >
            {toast.message}
          </Toast>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

function useToast() {
  const context = React.useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider')
  }
  return context
}

export { Toast, ToastProvider, useToast }
export type { ToastItem }
