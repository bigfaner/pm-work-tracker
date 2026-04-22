import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Pagination, PaginationPageSize } from './pagination'

describe('Pagination', () => {
  it('renders page buttons', () => {
    render(<Pagination currentPage={1} totalPages={5} onPageChange={vi.fn()} />)
    expect(screen.getByLabelText('Previous page')).toBeInTheDocument()
    expect(screen.getByLabelText('Next page')).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('highlights current page', () => {
    render(<Pagination currentPage={3} totalPages={5} onPageChange={vi.fn()} />)
    const activePage = screen.getByText('3')
    expect(activePage.className).toContain('bg-primary-50')
  })

  it('calls onPageChange when clicking a page', async () => {
    const onChange = vi.fn()
    render(<Pagination currentPage={1} totalPages={5} onPageChange={onChange} />)
    await userEvent.click(screen.getByText('3'))
    expect(onChange).toHaveBeenCalledWith(3)
  })

  it('disables prev on first page', () => {
    render(<Pagination currentPage={1} totalPages={5} onPageChange={vi.fn()} />)
    expect(screen.getByLabelText('Previous page')).toBeDisabled()
  })

  it('disables next on last page', () => {
    render(<Pagination currentPage={5} totalPages={5} onPageChange={vi.fn()} />)
    expect(screen.getByLabelText('Next page')).toBeDisabled()
  })

  it('shows ellipsis for large page counts', () => {
    render(<Pagination currentPage={5} totalPages={20} onPageChange={vi.fn()} />)
    const ellipses = screen.getAllByText('...')
    expect(ellipses.length).toBeGreaterThanOrEqual(1)
  })
})

describe('PaginationPageSize', () => {
  it('renders page size selector', () => {
    render(<PaginationPageSize pageSize={10} onPageSizeChange={vi.fn()} />)
    expect(screen.getByDisplayValue('10')).toBeInTheDocument()
    expect(screen.getByText('每页')).toBeInTheDocument()
    expect(screen.getByText('条')).toBeInTheDocument()
  })

  it('calls onPageSizeChange', async () => {
    const onChange = vi.fn()
    render(<PaginationPageSize pageSize={10} onPageSizeChange={onChange} />)
    await userEvent.selectOptions(screen.getByRole('combobox'), '20')
    expect(onChange).toHaveBeenCalledWith(20)
  })
})
