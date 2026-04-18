import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
  {
    variants: {
      variant: {
        // Status variants (7 item statuses)
        default: 'bg-slate-100 text-slate-600',
        'status-planning': 'bg-slate-100 text-slate-600',
        'status-in-progress': 'bg-blue-50 text-blue-700',
        'status-completed': 'bg-emerald-50 text-emerald-700',
        'status-on-hold': 'bg-amber-50 text-amber-800',
        'status-cancelled': 'bg-red-50 text-red-800',
        'status-overdue': 'bg-red-50 text-red-800',
        'status-pending': 'bg-yellow-50 text-yellow-800',
        // Priority variants (3 priority levels)
        'priority-high': 'bg-red-50 text-red-800',
        'priority-medium': 'bg-amber-50 text-amber-800',
        'priority-low': 'bg-slate-100 text-slate-600',
        // Semantic variants
        success: 'bg-emerald-50 text-emerald-700',
        warning: 'bg-amber-50 text-amber-800',
        error: 'bg-red-50 text-red-800',
        primary: 'bg-blue-50 text-blue-700',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <span className={cn(badgeVariants({ variant }), className)} {...props} />
  )
}

export { Badge, badgeVariants }
