import { describe, it, expect, vi, beforeEach, beforeAll, afterAll } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TooltipProvider } from '@/components/ui/tooltip'
import MilestoneTimeline from './MilestoneTimeline'
import type { Milestone, MilestoneMap } from '@/types'

// Mock API modules
vi.mock('@/api/milestones', () => ({
  listMilestonesByMapApi: vi.fn(),
  getMilestoneMapApi: vi.fn(),
  getMilestoneMapTransitionsApi: vi.fn(() => Promise.resolve([])),
  changeMilestoneMapStatusApi: vi.fn(),
  deleteMilestoneMapApi: vi.fn(),
  updateMilestoneMapApi: vi.fn(),
}))

vi.mock('@/api/mainItems', () => ({
  listMainItemsApi: vi.fn(),
  updateMainItemApi: vi.fn(),
}))

// Suppress console.error for DnD-related warnings in test env
const originalConsoleError = console.error
beforeAll(() => {
  console.error = (...args: unknown[]) => {
    if (typeof args[0] === 'string' && args[0].includes('onDragStart')) return
    originalConsoleError(...args)
  }
})
afterAll(() => {
  console.error = originalConsoleError
})

vi.mock('@/store/team', () => ({
  useTeamStore: vi.fn(
    (selector: (s: { currentTeamId: string | null }) => unknown) =>
      selector({ currentTeamId: 'team-1' }),
  ),
}))

vi.mock('@/components/ui/toast', () => ({
  useToast: vi.fn(() => ({ addToast: vi.fn() })),
}))

vi.mock('@/hooks/usePermission', () => ({
  usePermission: vi.fn(() => true),
}))

import { listMilestonesByMapApi, getMilestoneMapApi } from '@/api/milestones'
import { listMainItemsApi, updateMainItemApi } from '@/api/mainItems'
import { setDragMI } from './MilestoneTimeline'

const mockMap: MilestoneMap = {
  bizKey: 'map-1',
  teamKey: 'team-1',
  creatorKey: 'user-1',
  creatorName: '张三',
  assigneeKey: 'user-1',
  assigneeName: '张三',
  mapName: '产品 MVP',
  mapDesc: '产品首个可用版本，覆盖核心业务流程。',
  mapStatus: 'executing',
  statusName: '实施中',
  planStartDate: '2026-05-01',
  expectedEndDate: '2026-12-31',
  milestoneCount: 4,
  itemCount: 3,
  overallProgress: 60,
  milestoneSummary: [
    { bizKey: 'ms-1', name: '需求确认', status: 'completed', progress: 100 },
    { bizKey: 'ms-2', name: '开发', status: 'in_progress', progress: 60 },
    { bizKey: 'ms-3', name: '测试', status: 'not_started', progress: 0 },
    { bizKey: 'ms-4', name: '上线', status: 'not_started', progress: 0 },
  ],
  createTime: '2026-01-01',
  dbUpdateTime: '2026-01-01',
}

const mockMilestones: Milestone[] = [
  {
    bizKey: 'ms-1',
    teamKey: 'team-1',
    milestoneMapKey: 'map-1',
    milestoneName: 'MVP 发布',
    milestoneDesc: '',
    expectedEndDate: '2026-06-30',
    milestoneStatus: 'in_progress',
    statusName: '进行中',
    completion: 80,
    relatedMICount: 3,
    createTime: '2026-01-01',
    dbUpdateTime: '2026-01-01',
  },
  {
    bizKey: 'ms-2',
    teamKey: 'team-1',
    milestoneMapKey: 'map-1',
    milestoneName: 'Beta 测试',
    milestoneDesc: '',
    expectedEndDate: '2026-08-15',
    milestoneStatus: 'not_started',
    statusName: '未开始',
    completion: 0,
    relatedMICount: 0,
    createTime: '2026-01-01',
    dbUpdateTime: '2026-01-01',
  },
  {
    bizKey: 'ms-3',
    teamKey: 'team-1',
    milestoneMapKey: 'map-1',
    milestoneName: '正式上线',
    milestoneDesc: '',
    expectedEndDate: '2026-10-01',
    milestoneStatus: 'not_started',
    statusName: '未开始',
    completion: 0,
    relatedMICount: 0,
    createTime: '2026-01-01',
    dbUpdateTime: '2026-01-01',
  },
  {
    bizKey: 'ms-4',
    teamKey: 'team-1',
    milestoneMapKey: 'map-1',
    milestoneName: '迭代优化',
    milestoneDesc: '',
    expectedEndDate: '2026-11-15',
    milestoneStatus: 'cancelled',
    statusName: '已取消',
    completion: 0,
    relatedMICount: 0,
    createTime: '2026-01-01',
    dbUpdateTime: '2026-01-01',
  },
]

const mockMIs = {
  items: [
    {
      bizKey: 'mi-1',
      teamKey: 'team-1',
      code: 'MI-0001',
      title: '需求分析',
      priority: 'P1',
      proposerKey: 'user-1',
      assigneeKey: 'user-1',
      planStartDate: '2026-01-01',
      expectedEndDate: '2026-03-01',
      actualEndDate: null,
      itemStatus: 'progressing',
      statusName: '进行中',
      completion: 60,
      milestoneKey: 'ms-1',
      createTime: '2026-01-01',
      dbUpdateTime: '2026-01-01',
    },
    {
      bizKey: 'mi-2',
      teamKey: 'team-1',
      code: 'MI-0003',
      title: 'UI设计',
      priority: 'P1',
      proposerKey: 'user-1',
      assigneeKey: 'user-2',
      planStartDate: '2026-02-01',
      expectedEndDate: '2026-04-01',
      actualEndDate: '2026-04-01',
      itemStatus: 'completed',
      statusName: '已完成',
      completion: 100,
      milestoneKey: 'ms-1',
      createTime: '2026-01-15',
      dbUpdateTime: '2026-04-01',
    },
    {
      bizKey: 'mi-3',
      teamKey: 'team-1',
      code: 'MI-0005',
      title: 'API开发',
      priority: 'P2',
      proposerKey: 'user-1',
      assigneeKey: 'user-1',
      planStartDate: '2026-03-01',
      expectedEndDate: '2026-06-30',
      actualEndDate: null,
      itemStatus: 'progressing',
      statusName: '进行中',
      completion: 80,
      milestoneKey: 'ms-1',
      createTime: '2026-02-01',
      dbUpdateTime: '2026-02-01',
    },
  ],
  total: 3,
  page: 1,
  size: 200,
}

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderTimeline() {
  const onEditMap = vi.fn()
  const onEditMilestone = vi.fn()
  const onQuickAdd = vi.fn()

  vi.mocked(getMilestoneMapApi).mockResolvedValue(mockMap)
  vi.mocked(listMilestonesByMapApi).mockResolvedValue({
    items: mockMilestones,
    total: mockMilestones.length,
  })
  vi.mocked(listMainItemsApi).mockResolvedValue(mockMIs)

  const result = render(
    <QueryClientProvider client={createQueryClient()}>
      <TooltipProvider>
        <MemoryRouter>
          <MilestoneTimeline
            mapId="map-1"
            onEditMap={onEditMap}
            onEditMilestone={onEditMilestone}
            onQuickAdd={onQuickAdd}
          />
        </MemoryRouter>
      </TooltipProvider>
    </QueryClientProvider>,
  )

  return { onEditMap, onEditMilestone, onQuickAdd, ...result }
}

describe('MilestoneTimeline', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // AC-1: Timeline renders milestone nodes positioned by expected_end_date
  it('renders all milestone nodes', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    expect(screen.getByTestId('milestone-node-ms-2')).toBeInTheDocument()
    expect(screen.getByTestId('milestone-node-ms-3')).toBeInTheDocument()
    expect(screen.getByTestId('milestone-node-ms-4')).toBeInTheDocument()
  })

  it('renders node details: name, date, MI count, completion', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByText('MVP 发布')).toBeInTheDocument()
    })
    expect(screen.getByText('2026-06-30')).toBeInTheDocument()
    expect(screen.getByText('3 个事项')).toBeInTheDocument()
    // Multiple elements may contain "80%" (node completion + MI items)
    expect(screen.getAllByText('80%').length).toBeGreaterThanOrEqual(1)
  })

  it('renders breadcrumb with link back to list', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByText('里程碑图')).toBeInTheDocument()
    })
    // Breadcrumb link goes to /milestones
    const breadcrumbLink = screen.getByText('里程碑图').closest('a')
    expect(breadcrumbLink).toHaveAttribute('href', '/milestones')
  })

  it('renders map name in breadcrumb', async () => {
    renderTimeline()
    await waitFor(() => {
      // Map name appears in breadcrumb (last item) and title h1
      const breadcrumb = screen.getByText('里程碑图').closest('nav')
      expect(breadcrumb).toHaveTextContent('产品 MVP')
    })
  })

  // AC-2: Zoom controls
  it('renders zoom controls (compact/standard/relaxed)', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    expect(screen.getByTestId('zoom-compact')).toBeInTheDocument()
    expect(screen.getByTestId('zoom-standard')).toBeInTheDocument()
    expect(screen.getByTestId('zoom-relaxed')).toBeInTheDocument()
  })

  it('defaults to standard zoom', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    const standardBtn = screen.getByTestId('zoom-standard')
    expect(standardBtn.className).toContain('text-primary')
  })

  // AC-3: Name search with debounce
  it('renders search input for milestone name', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    expect(screen.getByPlaceholderText('搜索里程碑...')).toBeInTheDocument()
  })

  it('renders status filter with StatusTagFilter', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    // StatusTagFilter renders tag buttons for each status
    expect(screen.getByText('未开始')).toBeInTheDocument()
    expect(screen.getByText('进行中')).toBeInTheDocument()
  })

  it('renders reset and refresh buttons', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    expect(screen.getByTestId('reset-filters-btn')).toBeInTheDocument()
    expect(screen.getByTestId('refresh-btn')).toBeInTheDocument()
  })

  // AC-4: Click milestone node opens detail panel
  it('opens detail panel when milestone node is clicked', async () => {
    const user = userEvent.setup()
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('milestone-node-ms-1'))
    // Detail panel should open (dialog with milestone name)
    await waitFor(() => {
      expect(
        screen.getByRole('dialog', { name: /里程碑详情/ }),
      ).toBeInTheDocument()
    })
  })

  // AC-1: Basic info card shows map metadata
  it('renders basic info card with metadata', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByText('张三')).toBeInTheDocument()
    })
    expect(screen.getByText('2026-05-01')).toBeInTheDocument()
    expect(screen.getByText('2026-12-31')).toBeInTheDocument()
    // Multiple elements may contain "60%" (overall progress + MI items)
    expect(screen.getAllByText('60%').length).toBeGreaterThanOrEqual(1)
  })

  // Empty state
  it('shows empty state when no milestones', async () => {
    // Set mocks with empty milestones data
    vi.mocked(getMilestoneMapApi).mockResolvedValue(mockMap)
    vi.mocked(listMilestonesByMapApi).mockResolvedValue({
      items: [],
      total: 0,
    })
    vi.mocked(listMainItemsApi).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      size: 200,
    })

    render(
      <QueryClientProvider client={createQueryClient()}>
        <TooltipProvider>
          <MemoryRouter>
            <MilestoneTimeline
              mapId="map-1"
              onEditMap={vi.fn()}
              onEditMilestone={vi.fn()}
              onQuickAdd={vi.fn()}
            />
          </MemoryRouter>
        </TooltipProvider>
      </QueryClientProvider>,
    )

    await waitFor(() => {
      expect(screen.getByText('暂无里程碑')).toBeInTheDocument()
    })
  })

  // Loading state
  it('shows loading skeleton while fetching', () => {
    vi.mocked(getMilestoneMapApi).mockReturnValue(new Promise(() => {}))
    renderTimeline()
    expect(document.querySelector('.animate-pulse')).toBeInTheDocument()
  })

  // bug: nodes with same/close dates must not overlap
  it('bug: nodes with same date should not overlap', async () => {
    const sameDateMilestones: Milestone[] = [
      { ...mockMilestones[0], expectedEndDate: '2026-06-30' },
      { ...mockMilestones[1], expectedEndDate: '2026-06-30' },
      { ...mockMilestones[2], expectedEndDate: '2026-10-01' },
    ]

    vi.mocked(getMilestoneMapApi).mockResolvedValue(mockMap)
    vi.mocked(listMilestonesByMapApi).mockResolvedValue({
      items: sameDateMilestones,
      total: sameDateMilestones.length,
    })
    vi.mocked(listMainItemsApi).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      size: 200,
    })

    render(
      <QueryClientProvider client={createQueryClient()}>
        <TooltipProvider>
          <MemoryRouter>
            <MilestoneTimeline
              mapId="map-1"
              onEditMap={vi.fn()}
              onEditMilestone={vi.fn()}
              onQuickAdd={vi.fn()}
            />
          </MemoryRouter>
        </TooltipProvider>
      </QueryClientProvider>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })

    const node1 = screen.getByTestId('milestone-node-ms-1').parentElement!
    const node2 = screen.getByTestId('milestone-node-ms-2').parentElement!

    const left1 = parseFloat(node1.style.left)
    const left2 = parseFloat(node2.style.left)

    // Node width is 160px (w-40); left edges must differ by at least 160px
    expect(Math.abs(left1 - left2)).toBeGreaterThanOrEqual(160)
  })

  // bug: nodes should spread proportionally across width for sparse dates
  it('bug: spreads nodes proportionally by date across container width', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-4')).toBeInTheDocument()
    })

    // mockMilestones span Jun 30 → Nov 15 (wide range)
    // With proportional layout on default 800px container, last node left ≈ 640
    // With even spacing (4 × 184px), last node left ≈ 552
    const lastNode = screen.getByTestId('milestone-node-ms-4').parentElement!
    const lastLeft = parseFloat(lastNode.style.left)
    expect(lastLeft).toBeGreaterThan(580)
  })

  // bug: arrows should connect consecutive milestones
  it('bug: renders arrows between consecutive milestones', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    const arrows = document.querySelectorAll('[data-testid="timeline-arrow"]')
    expect(arrows.length).toBeGreaterThanOrEqual(1)
  })

  // Tick marks render
  it('renders tick marks container', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('tick-marks')).toBeInTheDocument()
    })
  })

  // MI layer: items displayed below milestone nodes
  it('renders MI items below milestone nodes in MI layer', async () => {
    renderTimeline()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
    })
    // MI items belonging to ms-1 should be visible in the MI layer
    expect(screen.getByText('MI-0001')).toBeInTheDocument()
    expect(screen.getByText('需求分析')).toBeInTheDocument()
    expect(screen.getByText('MI-0003')).toBeInTheDocument()
    expect(screen.getByText('MI-0005')).toBeInTheDocument()
  })

  // AC-5: Drag-and-drop MI rebinding
  describe('AC-5: Drag-and-drop MI rebinding', () => {
    it('shows target highlight when dragging MI over a different milestone node', async () => {
      renderTimeline()
      await waitFor(() => {
        expect(screen.getByTestId('milestone-node-ms-2')).toBeInTheDocument()
      })

      // Simulate drag data set (as if MI was dragged from detail panel)
      setDragMI({
        miBizKey: 'mi-1',
        miCode: 'MI-0001',
        sourceMilestoneKey: 'ms-1',
      })

      const nodeMs2 = screen.getByTestId('milestone-node-ms-2')
      // Fire dragOver event using testing-library fireEvent (triggers React handlers)
      fireEvent.dragOver(nodeMs2, {
        dataTransfer: { dropEffect: 'none' },
      })

      // Node should show highlight state (ring-primary-200 class)
      expect(nodeMs2.className).toContain('ring-primary-200')
    })

    it('calls updateMainItemApi with correct milestoneKey on drop', async () => {
      vi.mocked(updateMainItemApi).mockResolvedValue({
        bizKey: 'mi-1',
        teamKey: 'team-1',
        code: 'MI-0001',
        title: '需求分析',
        priority: 'P1',
        proposerKey: 'user-1',
        assigneeKey: 'user-1',
        planStartDate: '2026-01-01',
        expectedEndDate: '2026-03-01',
        actualEndDate: null,
        itemStatus: 'progressing',
        statusName: '进行中',
        completion: 60,
        milestoneKey: 'ms-2',
        createTime: '2026-01-01',
        dbUpdateTime: '2026-01-01',
      })

      renderTimeline()
      await waitFor(() => {
        expect(screen.getByTestId('milestone-node-ms-2')).toBeInTheDocument()
      })

      // Set drag data
      setDragMI({
        miBizKey: 'mi-1',
        miCode: 'MI-0001',
        sourceMilestoneKey: 'ms-1',
      })

      const nodeMs2 = screen.getByTestId('milestone-node-ms-2')
      // Fire drop event using testing-library fireEvent
      fireEvent.drop(nodeMs2)

      await waitFor(() => {
        expect(updateMainItemApi).toHaveBeenCalledWith(
          'team-1',
          'mi-1',
          { milestoneKey: 'ms-2' },
        )
      })
    })

    it('does not call API when dropping on same milestone', async () => {
      renderTimeline()
      await waitFor(() => {
        expect(screen.getByTestId('milestone-node-ms-1')).toBeInTheDocument()
      })

      // Set drag data with same source as target
      setDragMI({
        miBizKey: 'mi-1',
        miCode: 'MI-0001',
        sourceMilestoneKey: 'ms-1',
      })

      const nodeMs1 = screen.getByTestId('milestone-node-ms-1')
      fireEvent.drop(nodeMs1)

      // Should NOT call the API
      expect(updateMainItemApi).not.toHaveBeenCalled()
    })

    it('clears drag-over highlight on drag leave', async () => {
      renderTimeline()
      await waitFor(() => {
        expect(screen.getByTestId('milestone-node-ms-2')).toBeInTheDocument()
      })

      // Set drag state and trigger dragOver first
      setDragMI({
        miBizKey: 'mi-1',
        miCode: 'MI-0001',
        sourceMilestoneKey: 'ms-1',
      })

      const nodeMs2 = screen.getByTestId('milestone-node-ms-2')
      fireEvent.dragOver(nodeMs2, {
        dataTransfer: { dropEffect: 'none' },
      })
      expect(nodeMs2.className).toContain('ring-primary-200')

      // Now fire dragLeave
      fireEvent.dragLeave(nodeMs2)

      // Highlight should be gone
      expect(nodeMs2.className).not.toContain('ring-primary-200')
    })
  })
})
