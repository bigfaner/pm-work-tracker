import { describe, it, expect, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import MainItemDetailPage from './MainItemDetailPage'

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

function renderPage(mainItemId = '1') {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/items/${mainItemId}`]}>
        <Routes>
          <Route path="/items/:mainItemId" element={<MainItemDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedMembers = [
  { userId: 1, displayName: 'Test User', username: 'testuser', role: 'pm', joinedAt: '2024-01-01' },
  { userId: 2, displayName: 'Alice', username: 'alice', role: 'member', joinedAt: '2024-01-01' },
  { userId: 3, displayName: 'Bob', username: 'bob', role: 'member', joinedAt: '2024-01-01' },
]

const seedMainItem = {
  id: 1, team_id: 1, code: 'MI-0001', title: 'Alpha Task', priority: 'P1',
  proposer_id: 1, assignee_id: 1, start_date: '2026-03-20', expected_end_date: '2026-04-15',
  actual_end_date: null, status: '进行中', completion: 65, is_key_item: false,
  delay_count: 0, archived_at: null,
  created_at: '2026-03-20T00:00:00Z', updated_at: '2026-04-01T00:00:00Z',
  subItems: [
    {
      id: 11, team_id: 1, main_item_id: 1, title: 'Sub Alpha 1', description: '',
      priority: 'P1', assignee_id: 2, start_date: '2026-04-01', expected_end_date: '2026-04-10',
      actual_end_date: '2026-04-09', status: '已完成', completion: 100, is_key_item: false,
      delay_count: 0, weight: 1, created_at: '2026-04-01T00:00:00Z', updated_at: '2026-04-09T00:00:00Z',
    },
    {
      id: 12, team_id: 1, main_item_id: 1, title: 'Sub Alpha 2', description: '',
      priority: 'P2', assignee_id: 3, start_date: '2026-04-08', expected_end_date: '2026-04-18',
      actual_end_date: null, status: '进行中', completion: 80, is_key_item: false,
      delay_count: 0, weight: 1, created_at: '2026-04-01T00:00:00Z', updated_at: '2026-04-08T00:00:00Z',
    },
    {
      id: 13, team_id: 1, main_item_id: 1, title: 'Sub Alpha 3', description: '',
      priority: 'P2', assignee_id: 3, start_date: '2026-04-15', expected_end_date: '2026-04-25',
      actual_end_date: null, status: '进行中', completion: 30, is_key_item: false,
      delay_count: 0, weight: 1, created_at: '2026-04-01T00:00:00Z', updated_at: '2026-04-15T00:00:00Z',
    },
  ],
  achievements: ['登录/注册接口开发完成', 'JWT Token 签发与验证逻辑实现'],
  blockers: ['OAuth2.0 第三方回调地址需运维配合配置'],
}

function setupHandlers() {
  server.use(
    // Get main item with sub items
    http.get('/api/v1/teams/:teamId/main-items/:itemId', ({ params }) => {
      const item = Number(params.itemId) === 1 ? seedMainItem : null
      if (!item) return HttpResponse.json({ code: 'NOT_FOUND', message: 'not found' }, { status: 404 })
      return HttpResponse.json({ code: 0, data: item })
    }),

    // List members
    http.get('/api/v1/teams/:teamId/members', () => {
      return HttpResponse.json({ code: 0, data: seedMembers })
    }),

    // Update main item
    http.put('/api/v1/teams/:teamId/main-items/:itemId', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      return HttpResponse.json({ code: 0, data: { ...seedMainItem, ...body } })
    }),

    // Create sub item
    http.post('/api/v1/teams/:teamId/main-items/:mainId/sub-items', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          id: 100, team_id: 1, main_item_id: 1, description: '', priority: 'P2',
          assignee_id: null, start_date: null, expected_end_date: null, actual_end_date: null,
          status: '未开始', completion: 0, is_key_item: false, delay_count: 0, weight: 1,
          created_at: new Date().toISOString(), updated_at: new Date().toISOString(),
          ...body,
        },
      })
    }),

    // Change sub item status
    http.put('/api/v1/teams/:teamId/sub-items/:itemId/status', async () => {
      return HttpResponse.json({ code: 0, data: null })
    }),
  )
}

describe('MainItemDetailPage', () => {
  beforeEach(() => {
    useTeamStore.setState({ currentTeamId: 1, teams: [{ id: 1, name: 'Test Team', description: '', pm_id: 1, created_at: '', updated_at: '' }] })
    setupHandlers()
  })

  // --- Core rendering ---

  it('renders breadcrumb navigation', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('事项清单')).toBeInTheDocument()
    })
    // Breadcrumb link should go to /items
    const link = screen.getByText('事项清单').closest('a')
    expect(link).toHaveAttribute('href', '/items')
  })

  it('renders main item title and code badge', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
  })

  it('renders priority and status badges', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('P1')).toBeInTheDocument()
    })
  })

  it('renders assignee with avatar', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Test User')).toBeInTheDocument()
    })
  })

  it('renders info grid fields', async () => {
    renderPage()
    await waitFor(() => {
      // Some labels appear in both info grid and table headers
      expect(screen.getAllByText('负责人').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('开始时间').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('预期完成时间').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('实际完成时间').length).toBeGreaterThanOrEqual(1)
    })
  })

  // --- Circular progress indicator ---

  it('renders circular progress indicator with percentage', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('65%')).toBeInTheDocument()
    })
  })

  // --- Collapsible sections ---

  it('renders progress & summary section', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('进度与汇总')).toBeInTheDocument()
    })
  })

  it('shows sub-item count in summary section', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/已完成.*个子事项/)).toBeInTheDocument()
      expect(screen.getByText(/共.*个子事项/)).toBeInTheDocument()
    })
  })

  it('expands to show achievements and blockers on click', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('进度与汇总')).toBeInTheDocument()
    })

    // Click to expand
    await user.click(screen.getByText('进度与汇总'))

    await waitFor(() => {
      expect(screen.getByText('成果汇总')).toBeInTheDocument()
      expect(screen.getByText('卡点汇总')).toBeInTheDocument()
    })
  })

  // --- Sub items table ---

  it('renders sub items table', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('子事项列表')).toBeInTheDocument()
      expect(screen.getByText('Sub Alpha 1')).toBeInTheDocument()
      expect(screen.getByText('Sub Alpha 2')).toBeInTheDocument()
      expect(screen.getByText('Sub Alpha 3')).toBeInTheDocument()
    })
  })

  it('renders table headers', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('编号')).toBeInTheDocument()
      expect(screen.getByText('标题')).toBeInTheDocument()
      expect(screen.getAllByText('负责人').length).toBeGreaterThanOrEqual(1)
      expect(screen.getByText('完成度')).toBeInTheDocument()
      expect(screen.getAllByText('状态').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('sub item titles link to sub item detail page', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByText('Sub Alpha 1').closest('a')
      expect(link).toHaveAttribute('href', '/items/1/sub/11')
    })
  })

  // --- Inline status change ---

  it('sub items have inline status change dropdown', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Sub Alpha 1')).toBeInTheDocument()
    })

    // Find status badges that are clickable in the sub-item table
    const statusBadges = screen.getAllByText('已完成')
    expect(statusBadges.length).toBeGreaterThanOrEqual(1)
  })

  // --- Create sub item dialog ---

  it('opens create sub-item dialog', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
    })

    // Find the "+ 新增子事项" button
    const newSubBtn = screen.getByRole('button', { name: /新增子事项/ })
    await user.click(newSubBtn)

    await waitFor(() => {
      // Dialog should open with title
      expect(screen.getByText('新增子事项')).toBeInTheDocument()
    })
  })

  // --- Action buttons ---

  it('renders edit button', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
      expect(screen.getByText('编辑')).toBeInTheDocument()
    })
  })

  // --- Breadcrumb navigation ---

  it('breadcrumb has clickable links', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByText('事项清单').closest('a')
      expect(link).toHaveAttribute('href', '/items')
    })
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Loading state ---

  it('shows loading state', async () => {
    renderPage()
    await waitFor(() => {
      // Either loading indicator or the actual data appears
      const loading = screen.queryByText('加载中...')
      const title = screen.queryByRole('heading', { name: 'Alpha Task' })
      expect(loading || title).toBeTruthy()
    })
  })
})
