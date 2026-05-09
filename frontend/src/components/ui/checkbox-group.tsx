import { useCallback } from 'react'
import { cn } from '@/lib/utils'

export interface CheckboxOption {
  value: string
  label: string
}

interface CheckboxGroupProps {
  options: CheckboxOption[]
  selected: string[]
  onChange: (selected: string[]) => void
  title?: string
  disabled?: boolean
  className?: string
}

export function CheckboxGroup({
  options,
  selected,
  onChange,
  title,
  disabled = false,
  className,
}: CheckboxGroupProps) {
  const allSelected = options.length > 0 && options.every((o) => selected.includes(o.value))
  const someSelected = options.some((o) => selected.includes(o.value))
  const count = options.filter((o) => selected.includes(o.value)).length

  const toggleAll = useCallback(() => {
    if (allSelected) {
      // Remove all options in this group from selected
      const groupValues = new Set(options.map((o) => o.value))
      onChange(selected.filter((v) => !groupValues.has(v)))
    } else {
      // Add all options in this group to selected (deduplicated)
      const groupValues = options.map((o) => o.value)
      const merged = Array.from(new Set([...selected, ...groupValues]))
      onChange(merged)
    }
  }, [options, selected, onChange, allSelected])

  const toggleOne = useCallback(
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
    <div className={cn('space-y-1.5', className)}>
      {title && (
        <label className="flex items-center gap-2 text-sm font-medium text-primary cursor-pointer select-none">
          <input
            type="checkbox"
            checked={allSelected}
            ref={(el) => {
              if (el) el.indeterminate = someSelected && !allSelected
            }}
            onChange={toggleAll}
            disabled={disabled}
            className="w-4 h-4 rounded border-border-dark"
          />
          <span>{title}</span>
          <span className="text-xs text-tertiary font-normal">
            ({count}/{options.length})
          </span>
        </label>
      )}
      <div className="pl-6 space-y-1">
        {options.map((opt) => (
          <label
            key={opt.value}
            className="flex items-center gap-2 text-[13px] text-secondary cursor-pointer select-none"
          >
            <input
              type="checkbox"
              checked={selected.includes(opt.value)}
              onChange={() => toggleOne(opt.value)}
              disabled={disabled}
              className="w-4 h-4 rounded border-border-dark"
            />
            <span>{opt.label}</span>
          </label>
        ))}
      </div>
    </div>
  )
}
