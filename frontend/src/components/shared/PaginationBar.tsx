import { Pagination, PaginationPageSize } from '@/components/ui/pagination'

interface PaginationBarProps {
  currentPage: number
  totalPages: number
  total: number
  onPageChange: (page: number) => void
  pageSize?: number
  onPageSizeChange?: (size: number) => void
  pageSizeOptions?: number[]
}

export default function PaginationBar({
  currentPage,
  totalPages,
  total,
  onPageChange,
  pageSize,
  onPageSizeChange,
  pageSizeOptions = [10, 20, 50],
}: PaginationBarProps) {
  return (
    <div className="flex items-center justify-center gap-3 px-5 py-3 border-t border-border">
      <Pagination
        currentPage={currentPage}
        totalPages={Math.max(1, totalPages)}
        onPageChange={onPageChange}
      />
      {pageSize != null && onPageSizeChange != null && (
        <PaginationPageSize
          pageSize={pageSize}
          onPageSizeChange={onPageSizeChange}
          options={pageSizeOptions}
        />
      )}
      <span className="text-[13px] text-tertiary">共 {total} 条</span>
    </div>
  )
}
