import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'

interface SelectOption {
  value: string
  label: string
}

interface ItemFiltersProps {
  searchPlaceholder?: string
  searchValue?: string
  statusOptions: SelectOption[]
  statusValue?: string
  assigneeOptions: SelectOption[]
  assigneeValue?: string
  onSearchChange: (value: string) => void
  onStatusChange: (value: string) => void
  onAssigneeChange: (value: string) => void
  onReset: () => void
}

export default function ItemFilters({
  searchPlaceholder = '搜索...',
  searchValue = '',
  statusOptions,
  statusValue,
  assigneeOptions,
  assigneeValue,
  onSearchChange,
  onStatusChange,
  onAssigneeChange,
  onReset,
}: ItemFiltersProps) {
  return (
    <div className="flex items-center gap-3 mb-4 flex-wrap">
      <Input
        placeholder={searchPlaceholder}
        value={searchValue}
        onChange={(e) => onSearchChange(e.target.value)}
        className="w-[240px]"
      />
      <Select value={statusValue} onValueChange={onStatusChange}>
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="状态" />
        </SelectTrigger>
        <SelectContent>
          {statusOptions.map((opt) => (
            <SelectItem key={opt.value} value={opt.value || '_all'}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select value={assigneeValue} onValueChange={onAssigneeChange}>
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="负责人" />
        </SelectTrigger>
        <SelectContent>
          {assigneeOptions.map((opt) => (
            <SelectItem key={opt.value} value={opt.value || '_all'}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button variant="secondary" size="sm" onClick={onReset}>
        重置
      </Button>
    </div>
  )
}
