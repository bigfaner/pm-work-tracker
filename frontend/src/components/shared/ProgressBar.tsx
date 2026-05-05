import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";

interface ProgressBarProps {
  value: number;
  size?: "sm" | "default" | "lg";
  showPercentage?: boolean;
  className?: string;
}

function getIndicatorColor(value: number): string {
  if (value >= 100) return "bg-success";
  if (value > 0) return "bg-warning";
  return "bg-border";
}

export default function ProgressBar({
  value,
  size = "default",
  showPercentage = false,
  className,
}: ProgressBarProps) {
  const clampedValue = Math.max(0, Math.min(100, value));
  const indicatorColor = getIndicatorColor(clampedValue);

  return (
    <div className={cn("flex flex-col gap-1", className)}>
      <Progress
        value={clampedValue}
        size={size}
        indicatorClassName={indicatorColor}
      />
      {showPercentage && (
        <span className="text-xs text-tertiary block text-center">
          {Math.round(clampedValue)}%
        </span>
      )}
    </div>
  );
}
