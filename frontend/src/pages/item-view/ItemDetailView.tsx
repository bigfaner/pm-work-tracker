import { Fragment } from "react";
import { Link } from "react-router-dom";
import { Pencil, Plus } from "lucide-react";
import { MainItem, SubItem } from "@/types";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import PaginationBar from "@/components/shared/PaginationBar";
import PriorityBadge from "@/components/shared/PriorityBadge";
import StatusTransitionDropdown from "@/components/shared/StatusTransitionDropdown";
import { Badge } from "@/components/ui/badge";
import { PermissionGuard } from "@/components/PermissionGuard";
import { MAIN_ITEM_STATUSES, SUB_ITEM_STATUSES } from "@/lib/status";
import { isOverdue } from "@/lib/status";
import { showToast } from "@/lib/toast";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

function copyLink(path: string, title: string) {
  navigator.clipboard.writeText(`${window.location.origin}${path} ${title}`);
  showToast("链接已复制", "success");
}

function CodeBadge({
  label,
  path,
  title,
  className,
}: {
  label: string;
  path: string;
  title: string;
  className?: string;
}) {
  return (
    <TooltipProvider delayDuration={300}>
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            className={`font-mono cursor-pointer transition-colors ${className ?? ""}`}
            onClick={(e) => {
              e.stopPropagation();
              copyLink(path, title);
            }}
          >
            {label}
          </span>
        </TooltipTrigger>
        <TooltipContent>
          <p>点击复制链接</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

interface DetailViewProps {
  items: (MainItem & { subItems?: SubItem[] })[];
  subItemsMap: Record<string, SubItem[]>;
  memberName: (id: string | null) => string;
  formatDate: (date: string | null) => string;
  teamId: string;
  onRefresh: () => void;
  currentPage: number;
  totalPages: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  totalItems: number;
  onAddSubItem: (mainItemId: string, mainItemTitle: string) => void;
  onEditMainItem: (item: MainItem) => void;
  onAppendProgress: (
    subItemId: string,
    subItemTitle: string,
    subItemCompletion: number,
  ) => void;
  onEditSubItem: (sub: SubItem, mainItemBizKey: string) => void;
}

export default function ItemDetailView({
  items,
  subItemsMap,
  memberName,
  formatDate,
  teamId,
  onRefresh,
  currentPage,
  totalPages,
  pageSize,
  onPageChange,
  onPageSizeChange,
  totalItems,
  onAddSubItem,
  onAppendProgress,
  onEditSubItem,
}: DetailViewProps) {
  return (
    <div className="rounded-xl border border-border bg-white shadow-sm">
      <div data-testid="detail-table" className="overflow-x-auto">
        <Table className="min-w-[1100px]">
          <TableHeader>
            <TableRow>
              <TableHead className="whitespace-nowrap">编号</TableHead>
              <TableHead className="whitespace-nowrap">优先级</TableHead>
              <TableHead className="min-w-[180px]">标题</TableHead>
              <TableHead className="whitespace-nowrap">负责人</TableHead>
              <TableHead className="whitespace-nowrap">进度</TableHead>
              <TableHead className="whitespace-nowrap">状态</TableHead>
              <TableHead className="whitespace-nowrap">开始时间</TableHead>
              <TableHead className="whitespace-nowrap">预期完成时间</TableHead>
              <TableHead className="whitespace-nowrap">结束时间</TableHead>
              <TableHead className="whitespace-nowrap">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => {
              const subs = subItemsMap[item.bizKey];
              return (
                <Fragment key={item.bizKey}>
                  <TableRow className={subs?.length ? "bg-blue-50/40" : ""}>
                    <TableCell className="whitespace-nowrap">
                      <CodeBadge
                        label={item.code}
                        path={`/items/${item.bizKey}`}
                        title={item.title}
                        className="text-xs hover:bg-primary-100 hover:text-primary-600 px-0.5 rounded"
                      />
                    </TableCell>
                    <TableCell>
                      <PriorityBadge priority={item.priority} />
                    </TableCell>
                    <TableCell>
                      <Link
                        to={`/items/${item.bizKey}`}
                        className="font-medium text-primary-600 hover:text-primary-700 hover:underline truncate block max-w-xs"
                        title={item.title}
                      >
                        {item.title}
                      </Link>
                    </TableCell>
                    <TableCell className="whitespace-nowrap">
                      {memberName(item.assigneeKey)}
                    </TableCell>
                    <TableCell className="whitespace-nowrap">
                      <span className="text-xs">
                        {Math.round(item.completion)}%
                      </span>
                    </TableCell>
                    <TableCell className="whitespace-nowrap">
                      <StatusTransitionDropdown
                        currentStatus={item.itemStatus}
                        itemType="main"
                        teamId={teamId}
                        itemId={item.bizKey}
                        onStatusChanged={onRefresh}
                      />
                    </TableCell>
                    <TableCell className="text-xs whitespace-nowrap">
                      {formatDate(item.planStartDate)}
                    </TableCell>
                    <TableCell className="text-xs whitespace-nowrap">
                      <span>{formatDate(item.expectedEndDate)}</span>
                      {isOverdue(
                        item.expectedEndDate ?? undefined,
                        item.itemStatus,
                        new Date(),
                      ) && (
                        <Badge variant="error" className="ml-1">
                          延期
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-xs whitespace-nowrap">
                      {formatDate(item.actualEndDate)}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-0.5 whitespace-nowrap">
                        <Link to={`/items/${item.bizKey}`}>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-primary-600"
                            disabled={
                              !!MAIN_ITEM_STATUSES[
                                item.itemStatus as keyof typeof MAIN_ITEM_STATUSES
                              ]?.terminal
                            }
                          >
                            <Pencil size={14} />
                            编辑
                          </Button>
                        </Link>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-primary-600"
                          disabled={
                            !!MAIN_ITEM_STATUSES[
                              item.itemStatus as keyof typeof MAIN_ITEM_STATUSES
                            ]?.terminal
                          }
                          onClick={() => onAddSubItem(item.bizKey, item.title)}
                        >
                          <Plus size={14} />
                          添加子事项
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                  {subs?.map((sub) => (
                    <TableRow
                      key={`sub-${sub.bizKey}`}
                      className="bg-bg-alt/60"
                    >
                      <TableCell className="whitespace-nowrap">
                        <CodeBadge
                          label={sub.code.split("-").pop()!}
                          path={`/items/${item.bizKey}/sub/${sub.bizKey}`}
                          title={sub.title}
                          className="text-[11px] text-tertiary ml-4 hover:bg-primary-100 hover:text-primary-600 px-0.5 rounded"
                        />
                      </TableCell>
                      <TableCell>
                        <PriorityBadge priority={sub.priority} />
                      </TableCell>
                      <TableCell>
                        <Link
                          to={`/items/${item.bizKey}/sub/${sub.bizKey}`}
                          className="font-medium text-primary-600 hover:text-primary-700 hover:underline ml-4"
                        >
                          {sub.title}
                        </Link>
                      </TableCell>
                      <TableCell className="whitespace-nowrap">
                        {memberName(sub.assigneeKey)}
                      </TableCell>
                      <TableCell className="whitespace-nowrap">
                        <span className="text-xs">
                          {Math.round(sub.completion)}%
                        </span>
                      </TableCell>
                      <TableCell className="whitespace-nowrap">
                        <StatusTransitionDropdown
                          currentStatus={sub.itemStatus}
                          itemType="sub"
                          teamId={teamId}
                          itemId={sub.bizKey}
                          parentItemId={item.bizKey}
                          onStatusChanged={onRefresh}
                        />
                      </TableCell>
                      <TableCell className="text-xs whitespace-nowrap">
                        {formatDate(sub.planStartDate)}
                      </TableCell>
                      <TableCell className="text-xs whitespace-nowrap">
                        {formatDate(sub.expectedEndDate)}
                      </TableCell>
                      <TableCell className="text-xs whitespace-nowrap">
                        {formatDate(sub.actualEndDate)}
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-0.5 whitespace-nowrap">
                          <PermissionGuard code="sub_item:update">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-primary-600"
                              disabled={
                                !!SUB_ITEM_STATUSES[
                                  sub.itemStatus as keyof typeof SUB_ITEM_STATUSES
                                ]?.terminal
                              }
                              onClick={() => onEditSubItem(sub, item.bizKey)}
                            >
                              <Pencil size={14} />
                              编辑
                            </Button>
                          </PermissionGuard>
                          <PermissionGuard code="progress:update">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-primary-600"
                              disabled={
                                !!SUB_ITEM_STATUSES[
                                  sub.itemStatus as keyof typeof SUB_ITEM_STATUSES
                                ]?.terminal
                              }
                              onClick={() =>
                                onAppendProgress(
                                  sub.bizKey,
                                  sub.title,
                                  sub.completion,
                                )
                              }
                            >
                              <Plus size={14} />
                              追加进度
                            </Button>
                          </PermissionGuard>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </Fragment>
              );
            })}
          </TableBody>
        </Table>
      </div>
      <PaginationBar
        currentPage={currentPage}
        totalPages={totalPages}
        total={totalItems}
        onPageChange={onPageChange}
        pageSize={pageSize}
        onPageSizeChange={onPageSizeChange}
        pageSizeOptions={[5, 10, 20, 50]}
      />
    </div>
  );
}
