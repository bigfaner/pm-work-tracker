import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import MilestoneMapCard from './MilestoneMapCard'
import type { MilestoneMap } from '@/types'

function makeMap(overrides: Partial<MilestoneMap> = {}): MilestoneMap {
  return {
    bizKey: 'map-1',
    teamKey: 'team-1',
    creatorKey: 'user-1',
    creatorName: '张三',
    assigneeKey: 'user-1',
    assigneeName: '张三',
    mapName: '产品 MVP',
    mapDesc: '',
    mapStatus: 'executing',
    statusName: '实施中',
    planStartDate: '2026-05-01',
    expectedEndDate: '2026-12-31',
    milestoneCount: 4,
    itemCount: 12,
    overallProgress: 60,
    milestoneSummary: [
      { bizKey: 'ms-1', name: 'M1 需求确认', status: 'completed', progress: 100 },
      { bizKey: 'ms-2', name: 'M2 开发', status: 'in_progress', progress: 60 },
      { bizKey: 'ms-3', name: 'M3 测试', status: 'not_started', progress: 0 },
      { bizKey: 'ms-4', name: 'M4 上线', status: 'not_started', progress: 0 },
    ],
    createTime: '2026-01-01T00:00:00Z',
    dbUpdateTime: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function renderCard(map: MilestoneMap = makeMap()) {
  return render(
    <MemoryRouter>
      <MilestoneMapCard map={map} />
    </MemoryRouter>,
  )
}

describe('MilestoneMapCard', () => {
  // AC-1: Card renders with 4-line layout
  it('renders map name in row 1', () => {
    renderCard()
    expect(screen.getByText('产品 MVP')).toBeInTheDocument()
  })

  it('renders status badge in row 1', () => {
    renderCard()
    expect(screen.getByText('实施中')).toBeInTheDocument()
  })

  it('renders milestone count, item count, and assignee as separate items in row 2', () => {
    renderCard()
    expect(screen.getByText('4 个里程碑')).toBeInTheDocument()
    expect(screen.getByText('12 个事项')).toBeInTheDocument()
    expect(screen.getByText('张三')).toBeInTheDocument()
  })

  it('renders date span in row 3', () => {
    renderCard()
    expect(screen.getByText(/2026-05-01.*2026-12-31/)).toBeInTheDocument()
  })

  it('renders overall progress with percentage', () => {
    renderCard()
    expect(screen.getByText('整体进度')).toBeInTheDocument()
    // Multiple elements may contain "60%" (overall + milestone summary)
    expect(screen.getAllByText('60%').length).toBeGreaterThanOrEqual(1)
  })

  it('renders dot-and-line thumbnail with first and last milestone names in row 4', () => {
    renderCard()
    // First and last milestone names should be visible
    expect(screen.getByText('M1 需求确认')).toBeInTheDocument()
    expect(screen.getByText('M4 上线')).toBeInTheDocument()
    // Middle milestone names should NOT be rendered (only dots)
    expect(screen.queryByText('M2 开发')).not.toBeInTheDocument()
    expect(screen.queryByText('M3 测试')).not.toBeInTheDocument()
  })

  it('renders status-colored dots in row 4', () => {
    renderCard()
    const dots = screen.getAllByTestId('milestone-dot')
    expect(dots).toHaveLength(4)
    // completed → green (success)
    expect(dots[0]).toHaveClass('bg-success')
    // in_progress → blue (info)
    expect(dots[1]).toHaveClass('bg-info')
    // not_started → gray
    expect(dots[2]).toHaveClass('bg-gray-300')
  })

  it('renders connecting lines between dots in row 4', () => {
    renderCard()
    const lines = screen.getAllByTestId('milestone-line')
    // 4 milestones → 3 connecting lines
    expect(lines).toHaveLength(3)
  })

  it('shows "暂无里程碑" when milestoneCount is 0', () => {
    renderCard(makeMap({ milestoneCount: 0, milestoneSummary: [] }))
    expect(screen.getByText('暂无里程碑')).toBeInTheDocument()
  })

  // AC-4: Card click navigates to /milestones/:mapId
  it('links to /milestones/:mapId', () => {
    renderCard()
    const link = screen.getByTestId('milestone-map-card-map-1')
    expect(link).toHaveAttribute('href', '/milestones/map-1')
  })

  // AC-5: Hover effect class
  it('has hover effect classes', () => {
    renderCard()
    const link = screen.getByTestId('milestone-map-card-map-1')
    expect(link.className).toContain('hover:border-primary-300')
    expect(link.className).toContain('hover:shadow-md')
  })

  // Edge cases
  it('renders with null dates', () => {
    renderCard(makeMap({ planStartDate: null, expectedEndDate: null }))
    // No date span text should be present
    expect(screen.getByText('整体进度')).toBeInTheDocument()
  })

  it('renders with only start date', () => {
    renderCard(makeMap({ expectedEndDate: null }))
    expect(screen.getByText(/2026-05-01/)).toBeInTheDocument()
  })

  it('renders with only end date', () => {
    renderCard(makeMap({ planStartDate: null }))
    expect(screen.getByText(/2026-12-31/)).toBeInTheDocument()
  })
})
