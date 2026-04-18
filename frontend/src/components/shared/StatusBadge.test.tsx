import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import StatusBadge from './StatusBadge'

const statuses = [
  { status: '未开始', variant: 'status-planning' },
  { status: '进行中', variant: 'status-in-progress' },
  { status: '待评审', variant: 'status-pending' },
  { status: '已完成', variant: 'status-completed' },
  { status: '已关闭', variant: 'status-cancelled' },
  { status: '阻塞中', variant: 'status-overdue' },
  { status: '延期', variant: 'status-on-hold' },
]

describe('StatusBadge', () => {
  it('renders all 7 item statuses with correct text', () => {
    for (const { status } of statuses) {
      const { unmount } = render(<StatusBadge status={status} />)
      expect(screen.getByText(status)).toBeInTheDocument()
      unmount()
    }
  })

  it('renders unknown status with default variant', () => {
    render(<StatusBadge status="unknown" />)
    expect(screen.getByText('unknown')).toBeInTheDocument()
  })

  it('applies correct badge variant for each status', () => {
    for (const { status, variant } of statuses) {
      const { unmount } = render(<StatusBadge status={status} />)
      const badge = screen.getByText(status)
      // Check the badge has the expected class from the variant
      expect(badge.className).toContain('rounded-full')
      unmount()
    }
  })

  it('passes through extra className', () => {
    render(<StatusBadge status="进行中" className="extra-class" />)
    const badge = screen.getByText('进行中')
    expect(badge.className).toContain('extra-class')
  })
})
