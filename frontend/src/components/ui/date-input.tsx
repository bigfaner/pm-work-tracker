import { forwardRef } from 'react'
import { cn } from '@/lib/utils'

type DateInputProps = Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'>;

const DateInput = forwardRef<HTMLInputElement, DateInputProps>(
  ({ className, ...props }, forwardedRef) => {
    return (
      <div
        className={cn(
          'relative h-10 w-full rounded-md border border-border-dark bg-white shadow-sm transition-all duration-150',
          'focus-within:border-primary-500 focus-within:ring-2 focus-within:ring-primary-200',
          className,
        )}
      >
        <input
          type="date"
          ref={forwardedRef}
          className="h-full w-full cursor-pointer bg-transparent px-3 text-[13px] text-primary outline-none disabled:cursor-not-allowed disabled:opacity-50"
          {...props}
        />
      </div>
    )
  },
)
DateInput.displayName = 'DateInput'

export { DateInput }
