import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "../../lib/utils";
import {
  addWeeks,
  formatWeekLabel,
  getCurrentWeekStart,
} from "../../utils/weekUtils";

interface WeekPickerProps {
  weekStart: string;
  onChange: (weekStart: string) => void;
  maxWeek?: string;
  className?: string;
}

export function WeekPicker({
  weekStart,
  onChange,
  maxWeek,
  className,
}: WeekPickerProps) {
  const max = maxWeek ?? getCurrentWeekStart();
  const isAtMax = weekStart >= max;

  return (
    <div
      data-testid="week-selector"
      className={cn(
        "inline-flex items-center gap-1 h-8 rounded-md border border-border bg-white px-2",
        className,
      )}
    >
      <button
        aria-label="prev week"
        onClick={() => onChange(addWeeks(weekStart, -1))}
        className="flex items-center justify-center"
      >
        <ChevronLeft className="w-4 h-4" />
      </button>
      <span className="text-sm px-1 whitespace-nowrap">
        {formatWeekLabel(weekStart)}
      </span>
      <button
        aria-label="next week"
        disabled={isAtMax}
        onClick={() => {
          if (!isAtMax) onChange(addWeeks(weekStart, 1));
        }}
        className={cn(
          "flex items-center justify-center",
          isAtMax && "opacity-40 cursor-not-allowed",
        )}
      >
        <ChevronRight className="w-4 h-4" />
      </button>
    </div>
  );
}
