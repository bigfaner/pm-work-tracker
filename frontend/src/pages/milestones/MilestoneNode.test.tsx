import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TooltipProvider } from '@/components/ui/tooltip'
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
    <TooltipProvider>
      <MilestoneNode milestone={milestone} onClick={onClick} />
    </TooltipProvider>,
  )
  return { onClick, ...result }
}

describe('MilestoneNode', () => {
  // Card shows name + status badge + description (no dot, date, MI count)
  it('renders milestone name', () => {
    renderNode()
    expect(screen.getByText('MVP 发布')).toBeInTheDocument()
  })

  it('renders status badge', () => {
    renderNode()
    expect(screen.getByText('进行中')).toBeInTheDocument()
  })

  it('renders description', () => {
    renderNode()
    expect(screen.getByText('完成产品MVP')).toBeInTheDocument()
  })

  it('does not render date in card', () => {
    renderNode()
    expect(screen.queryByText('2026-06-30')).not.toBeInTheDocument()
  })

  it('does not render MI count in card', () => {
    renderNode()
    expect(screen.queryByText('3 个事项')).not.toBeInTheDocument()
  })

  it('does not render status dot', () => {
    const { container } = renderNode()
    const dot = container.querySelector('.text-primary.bg-current')
    expect(dot).not.toBeInTheDocument()
  })

  it('does not render completion percentage', () => {
    renderNode()
    expect(screen.queryByText('80%')).not.toBeInTheDocument()
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
    render(
      <TooltipProvider>
        <MilestoneNode milestone={makeMilestone()} selected={true} />
      </TooltipProvider>,
    )
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

  // Size: w-60 h-40
  it('has w-60 width class', () => {
    renderNode()
    const node = screen.getByTestId('milestone-node-ms-1')
    expect(node.className).toContain('w-60')
  })

  it('has h-40 fixed height class', () => {
    renderNode()
    const node = screen.getByTestId('milestone-node-ms-1')
    expect(node.className).toContain('h-40')
  })

  // DnD: isDragOver highlight
  it('shows drag-over highlight when isDragOver is true', () => {
    render(
      <TooltipProvider>
        <MilestoneNode
          milestone={makeMilestone()}
          isDragOver={true}
        />
      </TooltipProvider>,
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
      <TooltipProvider>
        <MilestoneNode
          milestone={makeMilestone()}
          onDragOver={onDragOver}
        />
      </TooltipProvider>,
    )
    const node = screen.getByTestId('milestone-node-ms-1')
    node.dispatchEvent(new Event('dragover', { bubbles: true }))
    expect(onDragOver).toHaveBeenCalled()
  })

  it('calls onDrop when provided', () => {
    const onDrop = vi.fn()
    render(
      <TooltipProvider>
        <MilestoneNode
          milestone={makeMilestone()}
          onDrop={onDrop}
        />
      </TooltipProvider>,
    )
    const node = screen.getByTestId('milestone-node-ms-1')
    node.dispatchEvent(new Event('drop', { bubbles: true }))
    expect(onDrop).toHaveBeenCalled()
  })
})
