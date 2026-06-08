import { useCallback } from 'react'
import { badgeVariants, type BadgeProps } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { getStatusVariant } from '@/lib/status'

interface StatusTagFilterProps {
  options: { value: string, label: string }[]
  selected: string[]
  onChange: (selected: string[]) => void
  label?: string
}

export function StatusTagFilter({
  options,
  selected,
  onChange,
  label = '状态',
}: StatusTagFilterProps) {
  const toggle = useCallback(
    (value: string) => {
      if (selected.includes(value)) {
        onChange(selected.filter((v) => v !== value))
      } else {
        onChange([...selected, value])
      }
    },
    [selected, onChange],
  )

  return (
    <div className="inline-flex items-center gap-2 rounded-md border border-border-dark bg-white px-2.5 py-1.5">
      <span className="text-xs text-tertiary shrink-0">{label}</span>
      <div className="w-px h-4 bg-border-dark shrink-0" />
      <div className="flex flex-wrap items-center gap-1.5">
        {options.map((opt) => {
        const isActive = selected.includes(opt.value)
        const variant = `status-${getStatusVariant(opt.value)}` as NonNullable<BadgeProps['variant']>

        return (
          <button
            key={opt.value}
            type="button"
            data-testid={`status-filter-${opt.value}`}
            onClick={() => toggle(opt.value)}
            className={cn(
              badgeVariants({ variant }),
              'cursor-pointer transition-all select-none',
              isActive
                ? 'ring-2 ring-offset-1 ring-current'
                : 'opacity-40 hover:opacity-70',
            )}
          >
            {opt.label}
          </button>
        )
      })}
      </div>
    </div>
  )
}
