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

  it('renders milestone and item counts with assignee in row 2', () => {
    renderCard()
    expect(screen.getByText('4 个里程碑 · 12 个事项')).toBeInTheDocument()
    expect(screen.getByText('张三')).toBeInTheDocument()
  })

  it('renders date span in row 3', () => {
    renderCard()
    expect(screen.getByText(/2026-05-01.*2026-12-31/)).toBeInTheDocument()
  })

  it('renders overall progress with percentage', () => {
    renderCard()
    expect(screen.getByText('整体进度')).toBeInTheDocument()
    expect(screen.getByText('60%')).toBeInTheDocument()
  })

  it('renders milestone node thumbnail dots in row 4', () => {
    const { container } = renderCard()
    // 4 milestones -> 4 dots
    const dots = container.querySelectorAll('.rounded-full.bg-primary-300')
    expect(dots.length).toBe(4)
  })

  it('shows "暂无里程碑" when milestoneCount is 0', () => {
    renderCard(makeMap({ milestoneCount: 0 }))
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
