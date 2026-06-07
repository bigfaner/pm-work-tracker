import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MilestoneNode from './MilestoneNode'
import type { Milestone } from '@/types'

function makeMilestone(overrides: Partial<Milestone> = {}): Milestone {
  return {
    bizKey: 'ms-1',
    teamKey: 'team-1',
    milestoneMapKey: 'map-1',
    milestoneName: 'MVP 发布',
    milestoneDesc: '完成产品MVP',
    expectedEndDate: '2026-06-30',
    milestoneStatus: 'in_progress',
    statusName: '进行中',
    completion: 80,
    relatedMICount: 3,
    createTime: '2026-01-01',
    dbUpdateTime: '2026-01-01',
    ...overrides,
  }
}

function renderNode(milestone: Milestone = makeMilestone()) {
  const onClick = vi.fn()
  const result = render(
    <MilestoneNode milestone={milestone} onClick={onClick} />,
  )
  return { onClick, ...result }
}

describe('MilestoneNode', () => {
  // AC-1: Node renders with status dot, name, completion %, date, MI count
  it('renders milestone name', () => {
    renderNode()
    expect(screen.getByText('MVP 发布')).toBeInTheDocument()
  })

  it('renders completion percentage', () => {
    renderNode()
    expect(screen.getByText('80%')).toBeInTheDocument()
  })

  it('renders expected end date', () => {
    renderNode()
    expect(screen.getByText('2026-06-30')).toBeInTheDocument()
  })

  it('renders MI count', () => {
    renderNode()
    expect(screen.getByText('3 个事项')).toBeInTheDocument()
  })

  it('renders status dot for in_progress', () => {
    const { container } = renderNode()
    const dot = container.querySelector('.text-primary.bg-current')
    expect(dot).toBeInTheDocument()
  })

  it('renders status dot for not_started with text-tertiary', () => {
    const { container } = renderNode(
      makeMilestone({ milestoneStatus: 'not_started' }),
    )
    const dot = container.querySelector('.text-tertiary.bg-current')
    expect(dot).toBeInTheDocument()
  })

  it('renders status dot for completed with text-success', () => {
    const { container } = renderNode(
      makeMilestone({ milestoneStatus: 'completed' }),
    )
    const dot = container.querySelector('.text-success.bg-current')
    expect(dot).toBeInTheDocument()
  })

  it('renders "未设置" when expectedEndDate is null', () => {
    renderNode(makeMilestone({ expectedEndDate: null }))
    expect(screen.getByText('未设置')).toBeInTheDocument()
  })

  // Click behavior
  it('calls onClick when clicked', async () => {
    const user = userEvent.setup()
    const { onClick } = renderNode()
    await user.click(screen.getByTestId('milestone-node-ms-1'))
    expect(onClick).toHaveBeenCalled()
  })

  it('calls onClick on Enter key', async () => {
    const user = userEvent.setup()
    const { onClick } = renderNode()
    const node = screen.getByTestId('milestone-node-ms-1')
    node.focus()
    await user.keyboard('{Enter}')
    expect(onClick).toHaveBeenCalled()
  })

  // Selected state
  it('shows selected styling when selected', () => {
    render(<MilestoneNode milestone={makeMilestone()} selected={true} />)
    const node = screen.getByTestId('milestone-node-ms-1')
    expect(node.className).toContain('border-primary')
    expect(node.className).toContain('ring-2')
  })

  it('shows hover styling when not selected', () => {
    renderNode()
    const node = screen.getByTestId('milestone-node-ms-1')
    expect(node.className).toContain('hover:bg-bg-alt')
  })

  // Cancelled state
  it('renders with opacity-50 when cancelled', () => {
    renderNode(makeMilestone({ milestoneStatus: 'cancelled' }))
    const node = screen.getByTestId('milestone-node-ms-1')
    expect(node.className).toContain('opacity-50')
  })

  // Size: w-40
  it('has w-40 width class', () => {
    renderNode()
    const node = screen.getByTestId('milestone-node-ms-1')
    expect(node.className).toContain('w-40')
  })

  // DnD: isDragOver highlight
  it('shows drag-over highlight when isDragOver is true', () => {
    render(
      <MilestoneNode
        milestone={makeMilestone()}
        isDragOver={true}
      />,
    )
    const node = screen.getByTestId('milestone-node-ms-1')
    expect(node.className).toContain('ring-primary-200')
    expect(node.className).toContain('bg-primary-50')
  })

  it('does not show drag-over highlight when isDragOver is false', () => {
    renderNode()
    const node = screen.getByTestId('milestone-node-ms-1')
    expect(node.className).not.toContain('ring-primary-200')
    expect(node.className).not.toContain('bg-primary-50')
  })

  it('calls onDragOver when provided', () => {
    const onDragOver = vi.fn()
    render(
      <MilestoneNode
        milestone={makeMilestone()}
        onDragOver={onDragOver}
      />,
    )
    const node = screen.getByTestId('milestone-node-ms-1')
    node.dispatchEvent(new Event('dragover', { bubbles: true }))
    expect(onDragOver).toHaveBeenCalled()
  })

  it('calls onDrop when provided', () => {
    const onDrop = vi.fn()
    render(
      <MilestoneNode
        milestone={makeMilestone()}
        onDrop={onDrop}
      />,
    )
    const node = screen.getByTestId('milestone-node-ms-1')
    node.dispatchEvent(new Event('drop', { bubbles: true }))
    expect(onDrop).toHaveBeenCalled()
  })
})
