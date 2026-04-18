import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ItemFilters from './ItemFilters'

describe('ItemFilters', () => {
  const defaultProps = {
    searchPlaceholder: '搜索标题...',
    statusOptions: [
      { value: '', label: '状态：全部' },
      { value: '未开始', label: '未开始' },
      { value: '进行中', label: '进行中' },
      { value: '已完成', label: '已完成' },
    ],
    assigneeOptions: [
      { value: '', label: '负责人：全部' },
      { value: '1', label: '张明' },
      { value: '2', label: '李华' },
    ],
    onSearchChange: vi.fn(),
    onStatusChange: vi.fn(),
    onAssigneeChange: vi.fn(),
    onReset: vi.fn(),
  }

  it('renders search input with placeholder', () => {
    render(<ItemFilters {...defaultProps} />)
    expect(screen.getByPlaceholderText('搜索标题...')).toBeInTheDocument()
  })

  it('renders status and assignee select triggers', () => {
    render(<ItemFilters {...defaultProps} />)
    // Radix Select uses buttons as triggers
    const triggers = screen.getAllByRole('combobox')
    expect(triggers.length).toBeGreaterThanOrEqual(2)
  })

  it('renders reset button', () => {
    render(<ItemFilters {...defaultProps} />)
    expect(screen.getByText('重置')).toBeInTheDocument()
  })

  it('calls onReset when reset button is clicked', async () => {
    const onReset = vi.fn()
    const user = userEvent.setup()
    render(<ItemFilters {...defaultProps} onReset={onReset} />)
    await user.click(screen.getByText('重置'))
    expect(onReset).toHaveBeenCalledOnce()
  })

  it('calls onSearchChange when typing in search input', async () => {
    const onSearchChange = vi.fn()
    const user = userEvent.setup()
    render(<ItemFilters {...defaultProps} onSearchChange={onSearchChange} />)
    const input = screen.getByPlaceholderText('搜索标题...')
    await user.type(input, '测试')
    expect(onSearchChange).toHaveBeenCalled()
  })
})
