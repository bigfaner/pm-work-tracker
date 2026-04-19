import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'

interface ProgressBarProps {
  value: number
  size?: 'sm' | 'default' | 'lg'
  showPercentage?: boolean
  className?: string
}

function getIndicatorColor(value: number): string {
  if (value >= 100) return 'bg-emerald-500'
  if (value > 0) return 'bg-amber-500'
  return 'bg-slate-200'
}

export default function ProgressBar({ value, size = 'default', showPercentage = false, className }: ProgressBarProps) {
  const clampedValue = Math.max(0, Math.min(100, value))
  const indicatorColor = getIndicatorColor(clampedValue)

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Progress value={clampedValue} size={size} indicatorClassName={indicatorColor} className="flex-1" />
      {showPercentage && (
        <span className="text-xs text-tertiary whitespace-nowrap">{clampedValue}%</span>
      )}
    </div>
  )
}
