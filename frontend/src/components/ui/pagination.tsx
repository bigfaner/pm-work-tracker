import * as React from 'react'
import { cn } from '@/lib/utils'

interface PaginationProps extends React.HTMLAttributes<HTMLElement> {
  currentPage: number
  totalPages: number
  onPageChange: (page: number) => void
}

function Pagination({
  currentPage,
  totalPages,
  onPageChange,
  className,
  ...props
}: PaginationProps) {
  const pages = getVisiblePages(currentPage, totalPages)

  return (
    <nav
      className={cn(
        'flex items-center justify-center gap-2 py-3 px-5 border-t border-border text-[13px] text-tertiary',
        className
      )}
      aria-label="Pagination"
      {...props}
    >
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage <= 1}
        className="inline-flex items-center justify-center w-8 h-8 rounded-md hover:bg-bg-alt disabled:opacity-50 disabled:pointer-events-none"
        aria-label="Previous page"
      >
        <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
          <path fillRule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clipRule="evenodd" />
        </svg>
      </button>

      {pages.map((page, idx) =>
        page === '...' ? (
          <span key={`ellipsis-${idx}`} className="px-1">
            ...
          </span>
        ) : (
          <button
            key={page}
            onClick={() => onPageChange(page as number)}
            className={cn(
              'inline-flex items-center justify-center w-8 h-8 rounded-md text-[13px]',
              currentPage === page
                ? 'bg-primary-50 text-primary-700 font-medium'
                : 'hover:bg-bg-alt'
            )}
          >
            {page}
          </button>
        )
      )}

      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage >= totalPages}
        className="inline-flex items-center justify-center w-8 h-8 rounded-md hover:bg-bg-alt disabled:opacity-50 disabled:pointer-events-none"
        aria-label="Next page"
      >
        <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
          <path fillRule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clipRule="evenodd" />
        </svg>
      </button>
    </nav>
  )
}

function getVisiblePages(current: number, total: number): (number | '...')[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)

  const pages: (number | '...')[] = [1]

  if (current > 3) pages.push('...')

  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)

  for (let i = start; i <= end; i++) pages.push(i)

  if (current < total - 2) pages.push('...')

  pages.push(total)
  return pages
}

interface PaginationPageSizeProps extends React.HTMLAttributes<HTMLDivElement> {
  pageSize: number
  onPageSizeChange: (size: number) => void
  options?: number[]
}

function PaginationPageSize({
  pageSize,
  onPageSizeChange,
  options = [10, 20, 50],
  className,
  ...props
}: PaginationPageSizeProps) {
  return (
    <div className={cn('flex items-center gap-2 text-[13px] text-tertiary', className)} {...props}>
      <span>Rows per page:</span>
      <select
        value={pageSize}
        onChange={(e) => onPageSizeChange(Number(e.target.value))}
        className="h-7 px-1 border border-border rounded-md text-[13px] bg-white text-primary cursor-pointer"
      >
        {options.map((opt) => (
          <option key={opt} value={opt}>
            {opt}
          </option>
        ))}
      </select>
    </div>
  )
}

export { Pagination, PaginationPageSize }
