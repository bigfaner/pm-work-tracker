import { Card, CardHeader, CardContent } from "@/components/ui/card";
import ProgressBar from "@/components/shared/ProgressBar";

interface ProgressSummaryCardProps {
  completion: number;
  completedCount: number;
  totalSubItems: number;
  expanded: boolean;
  onToggleExpanded: () => void;
  achievements?: string[];
  blockers?: string[];
}

export default function ProgressSummaryCard({
  completion,
  completedCount,
  totalSubItems,
  expanded,
  onToggleExpanded,
  achievements,
  blockers,
}: ProgressSummaryCardProps) {
  return (
    <Card className="mb-5">
      <CardHeader
        className="cursor-pointer select-none"
        onClick={onToggleExpanded}
      >
        <div className="flex items-center gap-4">
          {/* Circular progress SVG */}
          <svg width="56" height="56" viewBox="0 0 80 80">
            <circle
              cx="40"
              cy="40"
              r="34"
              fill="none"
              stroke="#e2e8f0"
              strokeWidth="6"
            />
            <circle
              cx="40"
              cy="40"
              r="34"
              fill="none"
              stroke="var(--color-primary-500, #3b82f6)"
              strokeWidth="6"
              strokeLinecap="round"
              strokeDasharray="213.6"
              strokeDashoffset={213.6 * (1 - completion / 100)}
              transform="rotate(-90 40 40)"
            />
            <text
              x="40"
              y="40"
              textAnchor="middle"
              dominantBaseline="central"
              fontSize="18"
              fontWeight="600"
              fill="var(--color-primary, #1e293b)"
            >
              {Math.round(completion)}%
            </text>
          </svg>
          <div>
            <h3 className="text-sm font-semibold text-primary m-0">
              进度与汇总
            </h3>
            <div className="text-[13px] text-secondary mt-0.5">
              已完成 {completedCount} 个子事项 / 共 {totalSubItems} 个子事项
            </div>
          </div>
        </div>
        <svg
          className={`w-4 h-4 text-tertiary transition-transform ${expanded ? "rotate-90" : ""}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth="2"
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      </CardHeader>
      {expanded && (
        <CardContent>
          <ProgressBar
            value={completion}
            size="default"
            showPercentage
            className="mb-5"
          />
          <div className="grid grid-cols-2 gap-5">
            <div>
              <div className="text-[13px] font-medium mb-2 text-success-text">
                成果汇总
              </div>
              <ul className="text-[13px] text-secondary pl-4 list-disc leading-relaxed">
                {achievements?.map((a, i) => <li key={i}>{a}</li>) || (
                  <li className="text-tertiary">暂无</li>
                )}
              </ul>
            </div>
            <div>
              <div className="text-[13px] font-medium mb-2 text-error-text">
                卡点汇总
              </div>
              <ul className="text-[13px] text-secondary pl-4 list-disc leading-relaxed">
                {blockers?.map((b, i) => <li key={i}>{b}</li>) || (
                  <li className="text-tertiary">暂无</li>
                )}
              </ul>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
}
