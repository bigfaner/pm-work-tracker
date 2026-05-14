import {
  describe,
  it,
  expect,
  vi,
  beforeEach,
  beforeAll,
  afterAll,
  afterEach,
} from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import { ToastProvider } from '@/components/ui/toast'
import { TooltipProvider } from '@/components/ui/tooltip'
import MilestoneDetailPanel from './MilestoneDetailPanel'
import type { Milestone, MainItem, PageResult } from '@/types'

// Mock usePermission
vi.mock('@/hooks/usePermission', () => ({
  usePermission: () => true,
}))

beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

function renderPanel(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <TooltipProvider>
          <MemoryRouter>{ui}</MemoryRouter>
        </TooltipProvider>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

const mockMilestone: Milestone = {
  bizKey: 'ms-1',
  teamKey: 'team-1',
  milestoneMapKey: 'map-1',
  milestoneName: 'M1 需求分析',
  expectedEndDate: '2026-06-30',
  milestoneStatus: 'in_progress',
  statusName: '进行中',
  completion: 60,
  relatedMICount: 3,
  createTime: '2026-05-01T00:00:00Z',
  dbUpdateTime: '2026-05-01T00:00:00Z',
}

const mockRelatedMIs: MainItem[] = [
  {
    bizKey: 'mi-1',
    teamKey: 'team-1',
    code: 'MI-0001',
    title: '需求分析',
    itemDesc: '',
    priority: 'P1',
    proposerKey: 'u-1',
    assigneeKey: 'u-1',
    planStartDate: '2026-05-01',
    expectedEndDate: '2026-06-30',
    actualEndDate: null,
    itemStatus: 'progressing',
    statusName: '进行中',
    completion: 60,
    milestoneKey: 'ms-1',
    createTime: '2026-05-01T00:00:00Z',
    dbUpdateTime: '2026-05-01T00:00:00Z',
  },
  {
    bizKey: 'mi-2',
    teamKey: 'team-1',
    code: 'MI-0003',
    title: 'UI设计',
    itemDesc: '',
    priority: 'P2',
    proposerKey: 'u-1',
    assigneeKey: 'u-1',
    planStartDate: '2026-05-01',
    expectedEndDate: '2026-06-30',
    actualEndDate: null,
    itemStatus: 'completed',
    statusName: '已完成',
    completion: 100,
    milestoneKey: 'ms-1',
    createTime: '2026-05-01T00:00:00Z',
    dbUpdateTime: '2026-05-01T00:00:00Z',
  },
]

function setupHandlers(milestone?: Milestone) {
  const ms = milestone || mockMilestone
  server.use(
    http.get('/v1/teams/:teamId/milestones/:bizKey', () => {
      return HttpResponse.json({ code: 0, data: ms })
    }),
    http.get('/v1/teams/:teamId/milestones/:bizKey/available-transitions', () => {
      return HttpResponse.json({ code: 0, data: { transitions: ['completed', 'cancelled'] } })
    }),
    http.get('/v1/teams/:teamId/main-items', () => {
      const page: PageResult<MainItem> = {
        items: mockRelatedMIs,
        total: mockRelatedMIs.length,
        page: 1,
        size: 100,
      }
      return HttpResponse.json({ code: 0, data: page })
    }),
  )
}

beforeEach(() => {
  useTeamStore.setState({ currentTeamId: 'team-1', teams: [] })
})

describe('MilestoneDetailPanel', () => {
  it('renders panel with milestone details', async () => {
    setupHandlers()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('M1 需求分析')).toBeInTheDocument()
    })
    expect(screen.getByText('2026-06-30')).toBeInTheDocument()
    expect(screen.getByText('进行中')).toBeInTheDocument()
  })

  it('shows loading skeleton initially', () => {
    setupHandlers()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    expect(screen.getByTestId('detail-panel')).toBeInTheDocument()
  })

  it('displays completion progress bar', async () => {
    setupHandlers()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('60%')).toBeInTheDocument()
    })
  })

  it('shows related MI list', async () => {
    setupHandlers()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText(/MI-0001/)).toBeInTheDocument()
    })
    expect(screen.getByText(/MI-0003/)).toBeInTheDocument()
  })

  it('shows add button when user has update permission', async () => {
    setupHandlers()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('M1 需求分析')).toBeInTheDocument()
    })

    expect(screen.getByText('添加')).toBeInTheDocument()
  })

  it('shows delete button when user has delete permission', async () => {
    setupHandlers()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('删除里程碑')).toBeInTheDocument()
    })
  })

  it('calls onClose when close button clicked', async () => {
    setupHandlers()
    const onClose = vi.fn()
    const user = userEvent.setup()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={onClose}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('M1 需求分析')).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: '关闭' }))
    expect(onClose).toHaveBeenCalled()
  })

  it('shows cancelled style for cancelled milestone', async () => {
    const cancelledMilestone: Milestone = {
      ...mockMilestone,
      milestoneStatus: 'cancelled',
      statusName: '已取消',
    }
    setupHandlers(cancelledMilestone)
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('已取消')).toBeInTheDocument()
    })
    expect(screen.getByText('里程碑已取消')).toBeInTheDocument()
  })

  it('opens delete confirmation dialog', async () => {
    setupHandlers()
    const user = userEvent.setup()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('删除里程碑')).toBeInTheDocument()
    })

    await user.click(screen.getByText('删除里程碑'))

    await waitFor(() => {
      expect(screen.getByText(/确定删除里程碑/)).toBeInTheDocument()
    })
    expect(screen.getByText('确认删除')).toBeInTheDocument()
  })

  it('calls unbind API when x button is clicked on MI row', async () => {
    setupHandlers()
    server.use(
      http.put('/v1/teams/:teamId/main-items/:bizKey', () => {
        return HttpResponse.json({ code: 0, data: {} })
      }),
    )
    const user = userEvent.setup()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText(/MI-0001/)).toBeInTheDocument()
    })

    // Find and click the unbind (x) button for the first MI
    const unbindButtons = screen.getAllByLabelText(/解绑/)
    await user.click(unbindButtons[0])
  })

  it('shows edit button when user has update permission', async () => {
    setupHandlers()
    renderPanel(
      <MilestoneDetailPanel
        milestoneBizKey="ms-1"
        onClose={() => {}}
        onDeleted={() => {}}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('M1 需求分析')).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: '编辑里程碑' })).toBeInTheDocument()
  })
})
