import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
  {
    variants: {
      variant: {
        // Status variants (7 item statuses)
        default: "bg-bg-alt text-secondary",
        "status-planning": "bg-bg-alt text-secondary",
        "status-in-progress": "bg-blue-50 text-blue-700",
        "status-completed": "bg-success-bg text-success-text",
        "status-on-hold": "bg-warning-bg text-warning-text",
        "status-cancelled": "bg-error-bg text-error-text",
        "status-overdue": "bg-error-bg text-error-text",
        "status-pending": "bg-yellow-50 text-yellow-800",
        // Priority variants (3 priority levels)
        "priority-high": "bg-error-bg text-error-text",
        "priority-medium": "bg-warning-bg text-warning-text",
        "priority-low": "bg-bg-alt text-secondary",
        // Semantic variants
        success: "bg-success-bg text-success-text",
        warning: "bg-warning-bg text-warning-text",
        error: "bg-error-bg text-error-text",
        primary: "bg-blue-50 text-blue-700",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

export interface BadgeProps
  extends
    React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <span className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
