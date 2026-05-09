import { forwardRef, useRef } from 'react'
import { Calendar } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface DateInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {}

const DateInput = forwardRef<HTMLInputElement, DateInputProps>(
  ({ className, ...props }, forwardedRef) => {
    const innerRef = useRef<HTMLInputElement | null>(null)

    const handleClick = () => {
      innerRef.current?.showPicker?.()
    }

    return (
      <div
        className={cn(
          'flex h-10 w-full items-center rounded-md border border-border-dark bg-white px-3 shadow-sm cursor-pointer transition-all duration-150',
          'focus-within:border-primary-500 focus-within:ring-2 focus-within:ring-primary-200',
          className
        )}
        onClick={handleClick}
      >
        <input
          type="date"
          ref={(node) => {
            innerRef.current = node
            if (typeof forwardedRef === 'function') forwardedRef(node)
            else if (forwardedRef) (forwardedRef as React.MutableRefObject<HTMLInputElement | null>).current = node
          }}
          className="flex-1 bg-transparent text-[13px] text-primary outline-none [&::-webkit-calendar-picker-indicator]:hidden disabled:cursor-not-allowed disabled:opacity-50"
          {...props}
        />
        <Calendar className="h-4 w-4 text-tertiary pointer-events-none shrink-0" />
      </div>
    )
  }
)
DateInput.displayName = 'DateInput'

export { DateInput }
