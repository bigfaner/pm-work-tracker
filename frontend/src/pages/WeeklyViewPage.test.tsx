import { describe, it, expect, beforeEach, beforeAll, afterAll, afterEach, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import WeeklyViewPage from './WeeklyViewPage'
import type { WeeklyViewResponse } from '@/types'

// Fix current week to a known date for deterministic tests
vi.mock('@/utils/weekUtils', async (importOriginal) => {
  const actual = await importOriginal() as typeof import('@/utils/weekUtils')
  return { ...actual, getCurrentWeekStart: () => '2026-04-13' }
})

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
      <MemoryRouter>
        <WeeklyViewPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const mockWeeklyResponse: WeeklyViewResponse = {
  weekStart: '2026-04-13',
  weekEnd: '2026-04-19',
  stats: {
    activeSubItems: 5,
    newlyCompleted: 2,
    inProgress: 3,
    blocked: 1,
    pending: 4,
    pausing: 0,
    overdue: 2,
  },
  groups: [
    {
      mainItem: {
        id: 1,
        code: 'M-001',
        title: '用户认证模块开发',
        priority: 'P1',
        status: '进行中',
        startDate: '2026-04-01',
        expectedEndDate: '2026-04-25',
        actualEndDate: null,
        completion: 58,
        subItemCount: 4,
      },
      lastWeek: [
        {
          id: 10,
          title: 'JWT Token 集成',
          priority: 'P2',
          status: 'progressing',
          assigneeName: '李伟',
          startDate: '2026-04-10',
          expectedEndDate: '2026-04-18',
          completion: 40,
          progressDescription: 'Token 签发逻辑开发中',
          progressRecords: [],
        },
        {
          id: 11,
          title: '权限中间件',
          priority: 'P2',
          status: 'progressing',
          assigneeName: '张明',
          startDate: '2026-04-12',
          expectedEndDate: '2026-04-21',
          completion: 20,
          progressDescription: '基础结构搭建中',
          progressRecords: [],
        },
      ],
      thisWeek: [
        {
          id: 10,
          title: 'JWT Token 集成',
          priority: 'P2',
          status: 'progressing',
          assigneeName: '李伟',
          startDate: '2026-04-10',
          expectedEndDate: '2026-04-18',
          completion: 70,
          progressDescription: 'Token 签发完成，黑名单联调中',
          progressRecords: [
            { id: 1, subItemId: 10, teamId: 1, authorId: 2, completion: 60, achievement: 'Token 签发完成', blocker: '', lesson: '', isPMCorrect: false, createdAt: '2026-04-15T10:00:00Z' },
            { id: 2, subItemId: 10, teamId: 1, authorId: 2, completion: 70, achievement: '黑名单联调中', blocker: 'Redis 连接超时', lesson: '', isPMCorrect: false, createdAt: '2026-04-17T14:00:00Z' },
          ],
          delta: 30,
          isNew: false,
          justCompleted: false,
        },
        {
          id: 11,
          title: '权限中间件',
          priority: 'P2',
          status: 'progressing',
          assigneeName: '张明',
          startDate: '2026-04-12',
          expectedEndDate: '2026-04-21',
          completion: 45,
          progressDescription: '中间件完成，RBAC 冲突待讨论',
          progressRecords: [
            { id: 3, subItemId: 11, teamId: 1, authorId: 3, completion: 45, achievement: '中间件完成', blocker: 'RBAC 冲突待讨论', lesson: '', isPMCorrect: false, createdAt: '2026-04-16T09:00:00Z' },
          ],
          delta: 25,
          isNew: false,
          justCompleted: false,
        },
        {
          id: 15,
          title: 'OAuth2 第三方登录',
          priority: 'P1',
          status: 'pending',
          assigneeName: '张明',
          startDate: '2026-04-20',
          expectedEndDate: '2026-04-25',
          completion: 0,
          progressDescription: '',
          progressRecords: [],
          delta: 0,
          isNew: true,
          justCompleted: false,
        },
      ],
      completedNoChange: [
        {
          id: 8,
          title: '登录页开发',
          priority: 'P1',
          status: 'completed',
          assigneeName: '王芳',
          startDate: '2026-04-01',
          expectedEndDate: '2026-04-10',
          completion: 100,
          progressDescription: '',
          progressRecords: [],
        },
      ],
    },
    {
      mainItem: {
        id: 2,
        code: 'M-002',
        title: '数据看板设计',
        priority: 'P2',
        status: '进行中',
        startDate: '2026-04-05',
        expectedEndDate: '2026-04-20',
        actualEndDate: null,
        completion: 80,
        subItemCount: 3,
      },
      lastWeek: [
        {
          id: 20,
          title: '数据看板前端',
          priority: 'P1',
          status: 'progressing',
          assigneeName: '刘洋',
          startDate: '2026-04-05',
          expectedEndDate: '2026-04-18',
          completion: 55,
          progressDescription: '基础组件开发中',
          progressRecords: [],
        },
      ],
      thisWeek: [
        {
          id: 20,
          title: '数据看板前端',
          priority: 'P1',
          status: 'completed',
          assigneeName: '刘洋',
          startDate: '2026-04-05',
          expectedEndDate: '2026-04-18',
          completion: 100,
          progressDescription: '图表渲染性能达标',
          progressRecords: [
            { id: 4, subItemId: 20, teamId: 1, authorId: 4, completion: 100, achievement: '图表渲染性能达标', blocker: '', lesson: '', isPMCorrect: false, createdAt: '2026-04-14T11:00:00Z' },
          ],
          delta: 45,
          isNew: false,
          justCompleted: true,
        },
        {
          id: 21,
          title: '看板数据接口',
          priority: 'P2',
          status: 'progressing',
          assigneeName: '陈刚',
          startDate: '2026-04-10',
          expectedEndDate: '2026-04-20',
          completion: 80,
          progressDescription: '接口开发完成，性能优化中',
          progressRecords: [],
          delta: 0,
          isNew: true,
          justCompleted: false,
        },
      ],
      completedNoChange: [],
    },
  ],
}

function setupWeeklyHandler(response = mockWeeklyResponse) {
  server.use(
    http.get('/v1/teams/:teamId/views/weekly', ({ request }) => {
      const url = new URL(request.url)
      const weekStart = url.searchParams.get('weekStart')
      if (!weekStart) {
        return HttpResponse.json(
          { code: 'VALIDATION_ERROR', message: 'weekStart is required' },
          { status: 400 },
        )
      }
      return HttpResponse.json({ code: 0, data: response })
    }),
  )
}

describe('WeeklyViewPage', () => {
  beforeEach(() => {
    useTeamStore.setState({
      currentTeamId: 1,
      teams: [{ id: 1, name: 'Test Team', description: '', pmId: 1, createdAt: '', updatedAt: '' }],
    })
    setupWeeklyHandler()
  })

  // --- Core rendering ---

  it('renders page with title', async () => {
    renderPage()
    expect(screen.getByText('每周进展')).toBeInTheDocument()
  })

  it('renders week selector (WeekPicker)', async () => {
    renderPage()
    expect(screen.getByTestId('week-selector')).toBeInTheDocument()
  })

  it('renders date range display', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/04\/13/)).toBeInTheDocument()
    })
  })

  // --- Stats bar ---

  it('renders all 4 stats with correct values', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('stat-active')).toHaveTextContent('5')
      expect(screen.getByTestId('stat-newly-completed')).toHaveTextContent('2')
      expect(screen.getByTestId('stat-in-progress')).toHaveTextContent('3')
      expect(screen.getByTestId('stat-blocked')).toHaveTextContent('1')
    })
  })

  it('renders stats labels', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('本周活跃')).toBeInTheDocument()
      // These labels appear in both stats bar and legend, use getAllByText
      expect(screen.getAllByText('本周新完成').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('进行中').length).toBeGreaterThanOrEqual(1)
      expect(screen.getByText('阻塞中')).toBeInTheDocument()
    })
  })

  it('renders newly completed stat in green', async () => {
    renderPage()
    await waitFor(() => {
      const stat2 = screen.getByTestId('stat-newly-completed')
      expect(stat2).toHaveClass('text-success-text')
    })
  })

  it('renders in progress stat in primary color', async () => {
    renderPage()
    await waitFor(() => {
      const stat3 = screen.getByTestId('stat-in-progress')
      expect(stat3).toHaveClass('text-primary-600')
    })
  })

  it('renders blocked stat in red', async () => {
    renderPage()
    await waitFor(() => {
      const stat4 = screen.getByTestId('stat-blocked')
      expect(stat4).toHaveClass('text-error')
    })
  })

  it('renders all 7 stat cards with correct values via testId', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('stat-active')).toHaveTextContent('5')
      expect(screen.getByTestId('stat-newly-completed')).toHaveTextContent('2')
      expect(screen.getByTestId('stat-in-progress')).toHaveTextContent('3')
      expect(screen.getByTestId('stat-blocked')).toHaveTextContent('1')
      expect(screen.getByTestId('stat-pending')).toHaveTextContent('4')
      expect(screen.getByTestId('stat-pausing')).toHaveTextContent('0')
      expect(screen.getByTestId('stat-overdue')).toHaveTextContent('2')
    })
  })

  it('renders pending/pausing/overdue labels', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('未开始')).toBeInTheDocument()
      expect(screen.getByText('暂停中')).toBeInTheDocument()
      expect(screen.getByText('逾期中')).toBeInTheDocument()
    })
  })

  it('shows tooltip content on click', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => expect(screen.getByTestId('stat-active')).toBeInTheDocument())
    const trigger = screen.getByTestId('stat-active').closest('button')!
    await user.click(trigger)
    await waitFor(() => {
      expect(screen.getByRole('tooltip')).toBeInTheDocument()
    })
  })

  it('stat card trigger has aria-describedby when tooltip is open', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => expect(screen.getByTestId('stat-active')).toBeInTheDocument())
    const trigger = screen.getByTestId('stat-active').closest('button')!
    await user.click(trigger)
    await waitFor(() => {
      expect(trigger).toHaveAttribute('aria-describedby')
    })
  })

  it('shows "-" for all stat cards when API returns error', async () => {
    server.use(
      http.get('/v1/teams/:teamId/views/weekly', () =>
        HttpResponse.json({ code: 'SERVER_ERROR', message: 'internal error' }, { status: 500 }),
      ),
    )
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('stat-active')).toHaveTextContent('-')
      expect(screen.getByTestId('stat-newly-completed')).toHaveTextContent('-')
      expect(screen.getByTestId('stat-in-progress')).toHaveTextContent('-')
      expect(screen.getByTestId('stat-blocked')).toHaveTextContent('-')
      expect(screen.getByTestId('stat-pending')).toHaveTextContent('-')
      expect(screen.getByTestId('stat-pausing')).toHaveTextContent('-')
      expect(screen.getByTestId('stat-overdue')).toHaveTextContent('-')
    })
  })

  // --- Comparison cards ---

  it('renders main item groups as comparison cards', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
      expect(screen.getByText('数据看板设计')).toBeInTheDocument()
    })
  })

  it('renders main item title as link to detail', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByText('用户认证模块开发').closest('a')
      expect(link).toHaveAttribute('href', '/items/1')
    })
  })

  it('renders priority badge in card header', async () => {
    renderPage()
    await waitFor(() => {
      const card1 = screen.getByTestId('group-card-1')
      // P1 appears in card header and sub-items, check at least one exists within the card
      const p1s = within(card1).getAllByText('P1')
      expect(p1s.length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders date range in card header', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/04\/01.*04\/25/)).toBeInTheDocument()
    })
  })

  it('renders sub-item count badge', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('4 个子事项')).toBeInTheDocument()
      expect(screen.getByText('3 个子事项')).toBeInTheDocument()
    })
  })

  it('renders progress bar in card header', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('58%')).toBeInTheDocument()
      expect(screen.getAllByText('80%').length).toBeGreaterThanOrEqual(1)
    })
  })

  // --- Two-column layout ---

  it('renders last week and this week column headers', async () => {
    renderPage()
    await waitFor(() => {
      // Each group has 上周 and 本周, so use getAllByText
      expect(screen.getAllByText(/上周/).length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText(/本周/).length).toBeGreaterThanOrEqual(2) // also in legend
    })
  })

  it('renders last week sub-items', async () => {
    renderPage()
    await waitFor(() => {
      // These appear in both last week and this week columns
      expect(screen.getAllByText('JWT Token 集成').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('权限中间件').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders this week sub-items', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('OAuth2 第三方登录')).toBeInTheDocument()
    })
  })

  it('renders sub-item status badges', async () => {
    renderPage()
    await waitFor(() => {
      const badges = screen.getAllByText('进行中')
      expect(badges.length).toBeGreaterThanOrEqual(2)
    })
  })

  it('renders sub-item assignee names', async () => {
    renderPage()
    await waitFor(() => {
      // Names may appear in multiple columns
      expect(screen.getAllByText('李伟').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('张明').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders sub-item progress records on separate lines', async () => {
    renderPage()
    await waitFor(() => {
      // Achievement text shown with prefix
      expect(screen.getByText(/成果：Token 签发完成/)).toBeInTheDocument()
      expect(screen.getByText(/成果：黑名单联调中/)).toBeInTheDocument()
      // Blocker text shown with prefix
      expect(screen.getByText(/卡点：Redis 连接超时/)).toBeInTheDocument()
      expect(screen.getByText(/卡点：RBAC 冲突待讨论/)).toBeInTheDocument()
    })
  })

  it('falls back to progressDescription when progressRecords is empty', async () => {
    // Override: use items with empty progressRecords but non-empty progressDescription
    setupWeeklyHandler({
      weekStart: '2026-04-13',
      weekEnd: '2026-04-19',
      stats: { activeSubItems: 1, newlyCompleted: 0, inProgress: 1, blocked: 0 },
      groups: [{
        mainItem: {
          id: 99,
          code: 'M-099',
          title: 'Fallback测试',
          priority: 'P2',
          status: '进行中',
          startDate: '2026-04-13',
          expectedEndDate: '2026-04-19',
          actualEndDate: null,
          completion: 50,
          subItemCount: 1,
        },
        lastWeek: [],
        thisWeek: [{
          id: 90,
          title: '回退测试子事项',
          priority: 'P2',
          status: 'progressing',
          assigneeName: '测试',
          startDate: '2026-04-13',
          expectedEndDate: '2026-04-19',
          completion: 50,
          progressDescription: '后端未更新时的描述',
          progressRecords: [],
          delta: 0,
          isNew: false,
          justCompleted: false,
        }],
        completedNoChange: [],
      }],
    })
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/50%.*后端未更新时的描述/)).toBeInTheDocument()
    })
  })

  // --- Delta badges ---

  it('renders +N% delta badge in green when progress increased', async () => {
    renderPage()
    await waitFor(() => {
      const deltaBadges = screen.getAllByText('+30%')
      expect(deltaBadges.length).toBeGreaterThanOrEqual(1)
      // Verify green styling
      const badge = deltaBadges[0]
      expect(badge).toHaveClass('text-success-text')
    })
  })

  it('renders completed check badge for justCompleted items', async () => {
    renderPage()
    await waitFor(() => {
      // justCompleted badge and legend both contain "完成 ✓"
      expect(screen.getAllByText('完成 ✓').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders NEW badge in amber for new items', async () => {
    renderPage()
    await waitFor(() => {
      const newBadges = screen.getAllByText('NEW')
      expect(newBadges.length).toBeGreaterThanOrEqual(1)
      const badge = newBadges[0]
      expect(badge).toHaveClass('text-warning-text')
    })
  })

  // --- Collapsed section ---

  it('collapses completed-no-change items by default', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
    })
    // The completed-no-change section should be hidden
    expect(screen.queryByText('已完成无变化')).not.toBeInTheDocument()
  })

  it('expands completed-no-change items when clicking expand button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
    })

    // Click expand button for first group
    const expandBtn = screen.getByTestId('expand-completed-1')
    await user.click(expandBtn)

    await waitFor(() => {
      expect(screen.getAllByText('已完成无变化').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('登录页开发').length).toBeGreaterThanOrEqual(1)
    })
  })

  // --- Week selector ---

  it('prevents selecting future weeks (next button disabled at max)', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('每周进展')).toBeInTheDocument()
    })
    const nextBtn = screen.getByLabelText('next week')
    expect(nextBtn).toBeDisabled()
  })

  // --- Completion percentage in sub-item rows (F7) ---

  it('shows completion percentage after title in sub-item row', async () => {
    renderPage()
    await waitFor(() => {
      // lastWeek item 10 has completion 40
      expect(screen.getAllByText('40%').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('applies text-secondary style when completion < 100', async () => {
    renderPage()
    await waitFor(() => {
      const pcts = screen.getAllByText('40%')
      expect(pcts[0]).toHaveClass('text-secondary')
      expect(pcts[0]).toHaveClass('font-semibold')
    })
  })

  it('applies text-success-text style when completion = 100', async () => {
    renderPage()
    await waitFor(() => {
      // thisWeek item 20 (数据看板前端) has completion 100
      const pcts = screen.getAllByText('100%')
      expect(pcts[0]).toHaveClass('text-success-text')
      expect(pcts[0]).toHaveClass('font-semibold')
    })
  })

  it('always shows completion percentage regardless of delta', async () => {
    renderPage()
    await waitFor(() => {
      // item 15 (OAuth2) has completion 0, isNew=true — still shows 0%
      expect(screen.getAllByText('0%').length).toBeGreaterThanOrEqual(1)
    })
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('每周进展')).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Empty state ---

  it('shows empty state when no groups returned', async () => {
    setupWeeklyHandler({
      weekStart: '2026-04-13',
      weekEnd: '2026-04-19',
      stats: { activeSubItems: 0, newlyCompleted: 0, inProgress: 0, blocked: 0 },
      groups: [],
    })
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/暂无周数据/)).toBeInTheDocument()
    })
  })


  // --- Week selector change ---

  it('prev week button navigates to previous week and refetches', async () => {
    const user = userEvent.setup()
    const requests: string[] = []
    server.use(
      http.get('/v1/teams/:teamId/views/weekly', ({ request }) => {
        const url = new URL(request.url)
        const weekStart = url.searchParams.get('weekStart') || ''
        requests.push(weekStart)
        return HttpResponse.json({ code: 0, data: mockWeeklyResponse })
      }),
    )
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
    })
    requests.length = 0

    await user.click(screen.getByLabelText('prev week'))

    await waitFor(() => {
      expect(requests.length).toBeGreaterThanOrEqual(1)
    })
    // Should be one week before getCurrentWeekStart (2026-04-13) = 2026-04-06
    expect(requests[requests.length - 1]).toBe('2026-04-06')
  })

  it('does not auto-generate SI- codes for sub-items', async () => {
    setupWeeklyHandler({
      ...mockWeeklyResponse,
      groups: [{
        ...mockWeeklyResponse.groups[0],
        thisWeek: [{
          id: 10,
          code: 'AUTH-00001-01',
          title: 'JWT Token 集成',
          priority: 'P2',
          status: 'progressing',
          assigneeName: '李伟',
          startDate: '2026-04-10',
          expectedEndDate: '2026-04-18',
          completion: 70,
          progressDescription: 'Token 签发完成',
          progressRecords: [],
          delta: 30,
          isNew: false,
          justCompleted: false,
        }],
        lastWeek: [],
        completedNoChange: [],
      }],
    })
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('JWT Token 集成')).toBeInTheDocument()
    })

    // Old auto-generated format should not appear anywhere in the DOM
    expect(document.body.textContent).not.toMatch(/SI-\d{3}-\d{2}/)
  })
})
