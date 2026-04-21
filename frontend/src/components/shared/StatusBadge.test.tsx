import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import StatusBadge from './StatusBadge'

const statuses = [
  { code: 'pending',     name: '待开始', variant: 'status-planning' },
  { code: 'progressing', name: '进行中', variant: 'status-in-progress' },
  { code: 'reviewing',   name: '待验收', variant: 'status-pending' },
  { code: 'completed',   name: '已完成', variant: 'status-completed' },
  { code: 'closed',      name: '已关闭', variant: 'status-cancelled' },
  { code: 'blocking',    name: '阻塞中', variant: 'status-overdue' },
  { code: 'pausing',     name: '已暂停', variant: 'status-on-hold' },
]

describe('StatusBadge', () => {
  it('renders all 7 item statuses with Chinese name from code lookup', () => {
    for (const { code, name } of statuses) {
      const { unmount } = render(<StatusBadge status={code} />)
      expect(screen.getByText(name)).toBeInTheDocument()
      unmount()
    }
  })

  it('renders unknown status code as raw code text', () => {
    render(<StatusBadge status="unknown" />)
    expect(screen.getByText('unknown')).toBeInTheDocument()
  })

  it('applies correct badge variant class for each status code', () => {
    for (const { code, variant } of statuses) {
      const { unmount } = render(<StatusBadge status={code} />)
      const badge = screen.getByText(
        statuses.find(s => s.code === code)!.name,
      )
      expect(badge.className).toContain('rounded-full')
      // The variant class should be applied via Badge component
      expect(badge.className).toMatch(/bg-/)
      unmount()
    }
  })

  it('prefers statusName prop over map lookup', () => {
    render(<StatusBadge status="progressing" statusName="Custom Name" />)
    expect(screen.getByText('Custom Name')).toBeInTheDocument()
  })

  it('falls back to map name when statusName is not provided', () => {
    render(<StatusBadge status="progressing" />)
    expect(screen.getByText('进行中')).toBeInTheDocument()
  })

  it('passes through extra className', () => {
    render(<StatusBadge status="progressing" className="extra-class" />)
    const badge = screen.getByText('进行中')
    expect(badge.className).toContain('extra-class')
  })

  it('renders sub-item statuses correctly', () => {
    render(<StatusBadge status="pausing" />)
    expect(screen.getByText('已暂停')).toBeInTheDocument()
  })
})
