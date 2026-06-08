import { forwardRef } from 'react'
import { Calendar } from 'lucide-react'
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
          className="h-full w-full cursor-pointer bg-transparent pl-3 pr-10 text-[13px] text-primary outline-none disabled:cursor-not-allowed disabled:opacity-50"
          {...props}
        />
        <Calendar className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-tertiary" />
      </div>
    )
  },
)
DateInput.displayName = 'DateInput'

export { DateInput }
