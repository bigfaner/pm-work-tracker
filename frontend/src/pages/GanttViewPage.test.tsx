import { describe, it, expect, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import { ToastProvider } from '@/components/ui/toast'
import GanttViewPage from './GanttViewPage'
import type { GanttViewResp } from '@/types'

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
        <MemoryRouter>
          <GanttViewPage />
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const mockGanttResponse: GanttViewResp = {
  items: [
    {
      id: 1,
      title: '用户认证模块开发',
      priority: 'P1',
      startDate: '2026-04-01',
      expectedEndDate: '2026-04-30',
      completion: 52,
      status: 'progressing',
      isOverdue: false,
      subItems: [
        {
          id: 10,
          title: '用户注册接口',
          startDate: '2026-04-01',
          expectedEndDate: '2026-04-15',
          completion: 60,
          status: 'progressing',
        },
        {
          id: 11,
          title: '登录鉴权实现',
          startDate: '2026-04-05',
          expectedEndDate: '2026-04-25',
          completion: 45,
          status: 'progressing',
        },
      ],
    },
    {
      id: 2,
      title: '数据报表系统',
      priority: 'P2',
      startDate: '2026-03-01',
      expectedEndDate: '2026-04-10',
      completion: 100,
      status: 'completed',
      isOverdue: false,
      subItems: [],
    },
    {
      id: 3,
      title: '性能优化项目',
      priority: 'P3',
      startDate: null,
      expectedEndDate: null,
      completion: 0,
      status: 'pending',
      isOverdue: false,
      subItems: [],
    },
    {
      id: 4,
      title: '移动端适配',
      priority: 'P2',
      startDate: '2026-03-15',
      expectedEndDate: '2026-04-10',
      completion: 30,
      status: 'progressing',
      isOverdue: true,
      subItems: [
        {
          id: 40,
          title: '响应式布局',
          startDate: '2026-03-15',
          expectedEndDate: '2026-04-08',
          completion: 50,
          status: 'progressing',
        },
      ],
    },
  ],
}

function setupGanttHandler(response = mockGanttResponse) {
  server.use(
    http.get('/v1/teams/:teamId/views/gantt', () => {
      return HttpResponse.json({ code: 0, data: response })
    }),
  )
}

describe('GanttViewPage', () => {
  beforeEach(() => {
    useTeamStore.setState({
      currentTeamId: 1,
      teams: [{ id: 1, name: 'Test Team', description: '', pmId: 1, createdAt: '', updatedAt: '' }],
    })
    setupGanttHandler()
  })

  // --- Core rendering ---

  it('renders page with data-testid', async () => {
    renderPage()
    expect(screen.getByTestId('gantt-view-page')).toBeInTheDocument()
  })

  it('renders page title', async () => {
    renderPage()
    expect(screen.getByText('整体进度')).toBeInTheDocument()
  })

  it('renders date range inputs', async () => {
    renderPage()
    expect(screen.getByTestId('date-start')).toBeInTheDocument()
    expect(screen.getByTestId('date-end')).toBeInTheDocument()
  })

  it('renders search input', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByPlaceholderText('搜索事项标题…')).toBeInTheDocument()
    })
  })

  // --- Gantt chart rendering ---

  it('renders gantt container', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('gantt-container')).toBeInTheDocument()
    })
  })

  it('renders label rows for each main item', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
      expect(screen.getByText('数据报表系统')).toBeInTheDocument()
      expect(screen.getByText('性能优化项目')).toBeInTheDocument()
      expect(screen.getByText('移动端适配')).toBeInTheDocument()
    })
  })

  it('renders timeline rows for each item', async () => {
    renderPage()
    await waitFor(() => {
      // Each main item + sub items gets a timeline row
      const timelineRows = screen.getAllByTestId(/^timeline-row-/)
      // 4 main items + 3 sub items = 7 rows
      expect(timelineRows.length).toBe(7)
    })
  })

  // --- Task bar status colors ---

  it('renders completed bar with green fill', async () => {
    renderPage()
    await waitFor(() => {
      const bar = screen.getByTestId('gantt-bar-2')
      expect(bar).toHaveClass('completed')
    })
  })

  it('renders overdue bar with red fill', async () => {
    renderPage()
    await waitFor(() => {
      const bar = screen.getByTestId('gantt-bar-4')
      expect(bar).toHaveClass('overdue')
    })
  })

  it('renders no-data bar with dashed border for items without dates', async () => {
    renderPage()
    await waitFor(() => {
      const bar = screen.getByTestId('gantt-bar-3')
      expect(bar).toHaveClass('no-data')
    })
  })

  it('renders in-progress bar with default blue fill', async () => {
    renderPage()
    await waitFor(() => {
      const bar = screen.getByTestId('gantt-bar-1')
      // No special class = default/in-progress
      expect(bar).not.toHaveClass('completed')
      expect(bar).not.toHaveClass('overdue')
      expect(bar).not.toHaveClass('no-data')
    })
  })

  // --- Today marker ---

  it('renders today marker line', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('gantt-today-line')).toBeInTheDocument()
    })
  })

  // --- Collapse/expand sub-items ---

  it('renders collapse toggle for items with sub-items', async () => {
    renderPage()
    await waitFor(() => {
      // Item 1 has sub-items, should have toggle
      const toggle1 = screen.getByTestId(`collapse-toggle-1`)
      expect(toggle1).toBeInTheDocument()
    })
  })

  it('hides toggle for items without sub-items', async () => {
    renderPage()
    await waitFor(() => {
      // Item 2 has no sub-items, toggle should be hidden
      const toggle2 = screen.getByTestId(`collapse-toggle-2`)
      expect(toggle2).toHaveAttribute('data-hidden', 'true')
    })
  })

  it('shows sub-item rows when clicking collapse toggle', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
    })

    // Sub-items should initially be hidden
    expect(screen.getByTestId('timeline-row-10')).toHaveClass('gantt-row-hidden')

    // Click toggle to expand
    const toggle = screen.getByTestId('collapse-toggle-1')
    await user.click(toggle)

    // Sub-items should now be visible
    await waitFor(() => {
      expect(screen.getByTestId('timeline-row-10')).not.toHaveClass('gantt-row-hidden')
      expect(screen.getByTestId('timeline-row-11')).not.toHaveClass('gantt-row-hidden')
    })
  })

  // --- Search/filter ---

  it('filters items by search keyword', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('搜索事项标题…')
    await user.type(searchInput, '认证')

    await waitFor(() => {
      // Should show matching item
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
      // Should hide non-matching items
      expect(screen.queryByText('数据报表系统')).not.toBeInTheDocument()
    })
  })

  // --- Load more ---

  it('renders load more button when there are more items', async () => {
    // Create response with more items than page size
    const manyItems: GanttViewResp = {
      items: Array.from({ length: 30 }, (_, i) => ({
        id: i + 1,
        title: `Item ${i + 1}`,
        priority: 'P2',
        startDate: '2026-04-01',
        expectedEndDate: '2026-04-30',
        completion: 50,
        status: 'progressing',
        isOverdue: false,
        subItems: [],
      })),
    }
    setupGanttHandler(manyItems)
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('load-more-btn')).toBeInTheDocument()
    })
  })

  it('hides load more button when all items loaded', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('用户认证模块开发')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('load-more-btn')).not.toBeInTheDocument()
  })

  it('loads more items on button click', async () => {
    const manyItems: GanttViewResp = {
      items: Array.from({ length: 30 }, (_, i) => ({
        id: i + 1,
        title: `Item ${i + 1}`,
        priority: 'P2',
        startDate: '2026-04-01',
        expectedEndDate: '2026-04-30',
        completion: 50,
        status: 'progressing',
        isOverdue: false,
        subItems: [],
      })),
    }
    setupGanttHandler(manyItems)
    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Item 1')).toBeInTheDocument()
    })

    // Initially shows first 20
    expect(screen.queryByText('Item 21')).not.toBeInTheDocument()

    // Click load more
    await user.click(screen.getByTestId('load-more-btn'))

    await waitFor(() => {
      expect(screen.getByText('Item 21')).toBeInTheDocument()
    })
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('整体进度')).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Empty state ---

  it('shows empty state when no items returned', async () => {
    setupGanttHandler({ items: [] })
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/暂无甘特图数据/)).toBeInTheDocument()
    })
  })

  // --- No team selected ---

  it('shows prompt when no team selected', async () => {
    useTeamStore.setState({ currentTeamId: null })
    renderPage()
    expect(screen.getByText('请先选择团队')).toBeInTheDocument()
  })

  // --- Percentage display ---

  it('renders completion percentage on bars', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('52%')).toBeInTheDocument()
      expect(screen.getByText('100%')).toBeInTheDocument()
    })
  })

  // --- No-data label ---

  it('renders "未设置时间" for items without dates', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('未设置时间')).toBeInTheDocument()
    })
  })

  // --- Layout structure matching prototype ---

  it('renders scroll-based layout with gantt-scroll and gantt-inner wrappers', async () => {
    renderPage()
    await waitFor(() => {
      const container = screen.getByTestId('gantt-container')
      expect(container).toHaveClass('gantt-container')
      const inner = container.querySelector('.gantt-inner')
      expect(inner).toBeInTheDocument()
    })
  })

  it('renders sticky label panel and scrollable timeline side by side', async () => {
    renderPage()
    await waitFor(() => {
      const container = screen.getByTestId('gantt-container')
      const labels = container.querySelector('.gantt-labels')
      expect(labels).toBeInTheDocument()
      const timeline = container.querySelector('.gantt-timeline')
      expect(timeline).toBeInTheDocument()
    })
  })

  // --- Weekday/weekend background colors ---

  it('renders weekday and weekend day cells', async () => {
    renderPage()
    await waitFor(() => {
      const container = screen.getByTestId('gantt-container')
      const weekdays = container.querySelectorAll('.gantt-day.weekday')
      const weekends = container.querySelectorAll('.gantt-day.weekend')
      expect(weekdays.length).toBeGreaterThan(0)
      expect(weekends.length).toBeGreaterThan(0)
    })
  })
})
