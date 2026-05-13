import {
  describe,
  it,
  expect,
  beforeEach,
  beforeAll,
  afterAll,
  afterEach,
  vi,
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
import MilestonesPage from './MilestonesPage'
import type { MilestoneMap, Milestone, PageResult } from '@/types'

// Mock usePermission to return true by default (full access)
vi.mock('@/hooks/usePermission', () => ({
  usePermission: () => true,
}))

// MSW lifecycle
beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderPage() {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <TooltipProvider>
          <MemoryRouter>
            <MilestonesPage />
          </MemoryRouter>
        </TooltipProvider>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const mockMaps: MilestoneMap[] = [
  {
    bizKey: 'map-1',
    teamKey: 'team-1',
    mapName: '产品 MVP',
    mapDesc: 'MVP version',
    mapStatus: 'executing',
    statusName: '实施中',
    milestoneCount: 4,
    itemCount: 8,
    overallProgress: 60,
    createTime: '2026-05-01T00:00:00Z',
    dbUpdateTime: '2026-05-01T00:00:00Z',
  },
  {
    bizKey: 'map-2',
    teamKey: 'team-1',
    mapName: '二期迭代',
    mapDesc: 'Phase 2',
    mapStatus: 'ready',
    statusName: '待实施',
    milestoneCount: 3,
    itemCount: 0,
    overallProgress: 0,
    createTime: '2026-05-01T00:00:00Z',
    dbUpdateTime: '2026-05-01T00:00:00Z',
  },
  {
    bizKey: 'map-3',
    teamKey: 'team-1',
    mapName: '技术债务清理',
    mapDesc: 'Tech debt',
    mapStatus: 'completed',
    statusName: '已完成',
    milestoneCount: 2,
    itemCount: 5,
    overallProgress: 100,
    createTime: '2026-05-01T00:00:00Z',
    dbUpdateTime: '2026-05-01T00:00:00Z',
  },
]

const mockMilestones: Milestone[] = [
  {
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
  },
  {
    bizKey: 'ms-2',
    teamKey: 'team-1',
    milestoneMapKey: 'map-1',
    milestoneName: 'M2 开发实现',
    expectedEndDate: '2026-07-31',
    milestoneStatus: 'not_started',
    statusName: '未开始',
    completion: 0,
    relatedMICount: 2,
    createTime: '2026-05-01T00:00:00Z',
    dbUpdateTime: '2026-05-01T00:00:00Z',
  },
]

function setupMapHandlers() {
  server.use(
    http.get('/v1/teams/:teamId/milestone-maps', ({ request }) => {
      const url = new URL(request.url)
      const status = url.searchParams.get('status')
      let filtered = mockMaps
      if (status) {
        filtered = mockMaps.filter((m) => m.mapStatus === status)
      }
      const page: PageResult<MilestoneMap> = {
        items: filtered,
        total: filtered.length,
        page: 1,
        size: 20,
      }
      return HttpResponse.json({ code: 0, data: page })
    }),
    http.post('/v1/teams/:teamId/milestone-maps', async ({ request }) => {
      const body = (await request.json()) as {
        mapName: string
        mapDesc?: string
      }
      const newMap: MilestoneMap = {
        bizKey: 'map-new',
        teamKey: 'team-1',
        mapName: body.mapName,
        mapDesc: body.mapDesc || '',
        mapStatus: 'planning',
        statusName: '规划中',
        milestoneCount: 0,
        itemCount: 0,
        overallProgress: 0,
        createTime: new Date().toISOString(),
        dbUpdateTime: new Date().toISOString(),
      }
      return HttpResponse.json({ code: 0, data: newMap })
    }),
    http.put('/v1/teams/:teamId/milestone-maps/:mapId', async ({ request }) => {
      const body = (await request.json()) as { mapName?: string, mapDesc?: string }
      const mapId = new URL(request.url).pathname.split('/').pop()
      const original = mockMaps.find((m) => m.bizKey === mapId) || mockMaps[0]
      return HttpResponse.json({
        code: 0,
        data: {
          ...original,
          ...(body.mapName && { mapName: body.mapName }),
          ...(body.mapDesc !== undefined && { mapDesc: body.mapDesc }),
        },
      })
    }),
    http.get('/v1/teams/:teamId/milestone-maps/:mapId/milestones', () => {
      return HttpResponse.json({ code: 0, data: { items: mockMilestones, total: mockMilestones.length } })
    }),
    http.get('/v1/teams/:teamId/milestone-maps/:mapId/available-transitions', () => {
      return HttpResponse.json({ code: 0, data: { transitions: ['reviewed', 'ready'] } })
    }),
  )
}

beforeEach(() => {
  useTeamStore.setState({ currentTeamId: 'team-1', teams: [] })
})

// --- Tests ---

describe('MilestonesPage', () => {
  describe('List View (first level)', () => {
    it('renders page title and create button', async () => {
      setupMapHandlers()
      renderPage()

      expect(screen.getByText('里程碑图')).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: /创建里程碑图/ }),
      ).toBeInTheDocument()
    })

    it('shows loading skeleton initially', () => {
      setupMapHandlers()
      renderPage()

      expect(screen.getByTestId('maps-loading')).toBeInTheDocument()
    })

    it('renders map cards after loading', async () => {
      setupMapHandlers()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })
      expect(screen.getByText('二期迭代')).toBeInTheDocument()
      expect(screen.getByText('技术债务清理')).toBeInTheDocument()
    })

    it('displays status badge on each card', async () => {
      setupMapHandlers()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('实施中')).toBeInTheDocument()
      })
      expect(screen.getByText('待实施')).toBeInTheDocument()
      expect(screen.getByText('已完成')).toBeInTheDocument()
    })

    it('displays milestone count and item count', async () => {
      setupMapHandlers()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('4 里程碑')).toBeInTheDocument()
      })
      expect(screen.getByText('8 事项')).toBeInTheDocument()
    })

    it('displays progress bar with percentage', async () => {
      setupMapHandlers()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('60%')).toBeInTheDocument()
      })
      expect(screen.getByText('100%')).toBeInTheDocument()
    })

    it('shows empty state when no maps', async () => {
      server.use(
        http.get('/v1/teams/:teamId/milestone-maps', () => {
          return HttpResponse.json({
            code: 0,
            data: { items: [], total: 0, page: 1, size: 20 },
          })
        }),
      )
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('暂无里程碑图')).toBeInTheDocument()
      })
    })

    it('shows error state when API fails', async () => {
      server.use(
        http.get('/v1/teams/:teamId/milestone-maps', () => {
          return HttpResponse.json(
            { code: 'INTERNAL_ERROR', message: 'internal error' },
            { status: 500 },
          )
        }),
      )
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('加载失败，请重试')).toBeInTheDocument()
      })
      expect(
        screen.getByRole('button', { name: /重试/ }),
      ).toBeInTheDocument()
    })

    it('filters maps by status', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      // Open status filter dropdown
      const filterTrigger = screen.getByTestId('status-filter-trigger')
      await user.click(filterTrigger)

      // Wait for dropdown to open, then select "已完成"
      const completedOption = await screen.findByRole('option', { name: '已完成' })
      await user.click(completedOption)

      await waitFor(() => {
        expect(screen.queryByText('产品 MVP')).not.toBeInTheDocument()
      })
      expect(screen.getByText('技术债务清理')).toBeInTheDocument()
    })

    it('refreshes list when refresh button is clicked', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      const refreshBtn = screen.getByRole('button', { name: /刷新/ })
      await user.click(refreshBtn)

      // Should still show data (re-fetched)
      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })
    })

    it('renders create button that respects milestone:create permission', async () => {
      setupMapHandlers()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      // usePermission is mocked to return true, so button should be enabled
      const createBtn = screen.getByRole('button', { name: /创建里程碑图/ })
      expect(createBtn).toBeEnabled()
    })
  })

  describe('Create Map Dialog (UF-7)', () => {
    it('opens create dialog when create button clicked', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /创建里程碑图/ }))

      // Dialog should have a title and input
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('请输入里程碑图名称')).toBeInTheDocument()
    })

    it('creates a new map via dialog', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /创建里程碑图/ }))

      const nameInput = screen.getByPlaceholderText('请输入里程碑图名称')
      await user.type(nameInput, '新里程碑图')

      const confirmBtn = screen.getByRole('button', { name: /确认/ })
      await user.click(confirmBtn)

      // Dialog should close
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })

    it('validates map name is required', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /创建里程碑图/ }))

      const confirmBtn = screen.getByRole('button', { name: /确认/ })
      await user.click(confirmBtn)

      // Should show validation error
      await waitFor(() => {
        expect(screen.getByText(/请输入名称/)).toBeInTheDocument()
      })
    })
  })

  describe('Timeline View (second level)', () => {
    it('enters timeline view when card is clicked', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      // Click on a map card
      await user.click(screen.getByText('产品 MVP'))

      // Should show breadcrumb
      await waitFor(() => {
        expect(screen.getByText('里程碑图')).toBeInTheDocument()
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      // Should show zoom controls
      expect(screen.getByRole('button', { name: '周' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: '月' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: '季' })).toBeInTheDocument()
    })

    it('shows milestone nodes in timeline', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      await user.click(screen.getByText('产品 MVP'))

      await waitFor(() => {
        expect(screen.getByText('M1 需求分析')).toBeInTheDocument()
      })
      expect(screen.getByText('M2 开发实现')).toBeInTheDocument()
    })

    it('shows empty timeline when no milestones', async () => {
      setupMapHandlers()
      server.use(
        http.get('/v1/teams/:teamId/milestone-maps/:mapId/milestones', () => {
          return HttpResponse.json({ code: 0, data: { items: [], total: 0 } })
        }),
      )

      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      await user.click(screen.getByText('产品 MVP'))

      await waitFor(() => {
        expect(screen.getByText('暂无里程碑')).toBeInTheDocument()
      })
    })

    it('navigates back to list via breadcrumb', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      await user.click(screen.getByText('产品 MVP'))

      await waitFor(() => {
        expect(screen.getByText('M1 需求分析')).toBeInTheDocument()
      })

      // Click back button
      await user.click(screen.getByRole('button', { name: /返回列表/ }))

      // Should show list view again
      await waitFor(() => {
        expect(screen.getByText('二期迭代')).toBeInTheDocument()
      })
    })

    it('switches zoom level', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      await user.click(screen.getByText('产品 MVP'))

      await waitFor(() => {
        expect(screen.getByText('M1 需求分析')).toBeInTheDocument()
      })

      // Click month zoom
      await user.click(screen.getByRole('button', { name: '月' }))

      // Timeline should still show milestones
      expect(screen.getByText('M1 需求分析')).toBeInTheDocument()
    })
  })

  describe('Status filter options', () => {
    it('shows all status filter options', async () => {
      setupMapHandlers()
      const user = userEvent.setup()
      renderPage()

      await waitFor(() => {
        expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      })

      const filterTrigger = screen.getByTestId('status-filter-trigger')
      await user.click(filterTrigger)

      await waitFor(() => {
        expect(screen.getByRole('option', { name: '规划中' })).toBeInTheDocument()
      })
      expect(screen.getByRole('option', { name: '已评审' })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: '待实施' })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: '实施中' })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: '已完成' })).toBeInTheDocument()
    })
  })
})
